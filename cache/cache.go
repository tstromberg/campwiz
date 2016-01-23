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
	// POST form values
	Form url.Values
	// Maximum age of content.
	MaxAge time.Duration
}

// Key returns a cache-key.
func (r Request) Key() []byte {
	return []byte(r.Method + "-" + r.URL + "?" + r.Form.Encode())
}

// Result defines which data may be cached for an HTTP response.
type Result struct {
	// Status Code
	StatusCode int

	// HTTP headers
	Header http.Header

	// Body is the entire HTTP message body.
	Body []byte
	// MTime is when this value was last updated in the cache.
	MTime time.Time
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
		return res, nil
	}

	client := &http.Client{Jar: cookieJar}
	hr, err := http.NewRequest(req.Method, req.URL, nil)
	if err != nil {
		return res, err
	}
	hr.Form = req.Form
	hr.Header.Add("User-Agent", userAgent)
	r, err := client.Do(hr)
	if err != nil {
		return res, err
	}

	// Write the response into the cache. Mask over any failures.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return Result{}, err
	}
	cr := Result{r.StatusCode, r.Header, body, time.Now()}

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
