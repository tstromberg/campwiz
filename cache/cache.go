// The autocamper package contains all of the brains for querying campsites.
package cache

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"

	"github.com/steveyen/gkvlite"
)

var (
	// cachePath is the location of the cache
	cachePath = os.ExpandEnv("${HOME}/.autocamper.cache")

	// store is a gkvlite Store
	store = getCacheStore()

	// collection is a gkvlite collection that documents can be fetched from.
	collection = store.SetCollection("cache", nil)

	// user-agent to emulate
	userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.48 Safari/537.36"

	// cookie jar
	cookieJar, _ = cookiejar.New(nil)
)


// Request defines what can be passed in as a request
type Request struct {
	// Method type
	Method string
	// URL
	URL string
	// Referrer
	Referrer string
	// Cookies
	Cookies []*http.Cookie
	// POST form values
	Form url.Values
	// Maximum age of content.
	MaxAge time.Duration
}

// Key returns a cache-key.
func (r Request) Key() []byte {
	var buf bytes.Buffer
	buf.WriteString(r.Method + " ")
	buf.WriteString(r.URL + "?" + r.Form.Encode())
	for _, c := range(r.Cookies) {
		buf.WriteString(fmt.Sprintf("+cookie=%s", c.String()))
	}
	return buf.Bytes()
}

// Result defines which data may be cached for an HTTP response.
type Result struct {
	// URL result is from
	URL string
	// Status Code
	StatusCode int
	// HTTP headers
	Header http.Header
	// Cookies are the cookies that came with the request.
	Cookies []*http.Cookie
	// Body is the entire HTTP message body.
	Body []byte
	// MTime is when this value was last updated in the cache.
	MTime time.Time
	// If entry was served from cache
	Cached bool
}

// tryCache attempts a cache-only fetch.
func tryCache(req Request) (Result, error) {
	var res Result
	cachedBytes, err := collection.Get(req.Key())
	if err != nil {
		return res, err
	}

	// Item is in cache, but we do not yet know if it is too old.
	buf := bytes.NewBuffer(cachedBytes)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&res)
	// Invalid item in cache?
	if err != nil {
		return res, err
	}

	age := time.Since(res.MTime)
	log.Printf("Item age is %s", age)
	if age > req.MaxAge {
		return res, fmt.Errorf("%s was too old.", req.URL)
	}
	return res, nil
}


// Fetch wraps http.Get/http.Post behind a persistent cache.
func Fetch(req Request) (Result, error) {
	res, err := tryCache(req)
	if err != nil {
		log.Printf("MISS[%s]: %v", req.Key(), req, err)
	} else {
		log.Printf("HIT[%s]: max-age: %d", req.Key(), req.MaxAge)
		res.Cached = true
		return res, nil
	}

	client := &http.Client{Jar: cookieJar}
	hr, err := http.NewRequest(req.Method, req.URL, bytes.NewBufferString(req.Form.Encode()))
	if err != nil {
		return res, err
	}

	hr.Header.Add("User-Agent", userAgent)

	for _, c := range(req.Cookies) {
		hr.AddCookie(c)
	}

	if req.Method == "POST" {
		hr.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	log.Printf("Fetching: %+v", hr)
	r, err := client.Do(hr)
	if err != nil {
		return res, err
	}

	// Write the response into the cache. Mask over any failures.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return Result{}, err
	}
	cr := Result{
		URL: req.URL,
		StatusCode: r.StatusCode,
		Header: r.Header,
		Cookies: r.Cookies(),
		Body: body,
		MTime: time.Now(),
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(&cr)
	if err != nil {
		return cr, fmt.Errorf("encoding %+v: %v", cr, err)
	} else {
		bufBytes, err := ioutil.ReadAll(&buf)
		log.Printf("Buf bytes: %s", bufBytes)
		if err != nil {
			log.Printf("Failed to read back encoded response: %s", err)
		} else {
			log.Printf("Storing %s", req.Key())
			collection.Set(req.Key(), bufBytes)
			store.Flush()
		}
	}
	cr.Cached = false
	return cr, nil
}

// Returns a gkvlite collection
func getCacheStore() *gkvlite.Store {
	f, err := os.OpenFile(cachePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatal(err)
	}
	s, err := gkvlite.NewStore(f)
	if err != nil {
		log.Fatal(err)
	}
	return s
}
