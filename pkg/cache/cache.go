// The campwiz package contains all of the brains for querying campsites.
package cache

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/moul/http2curl"
	"github.com/peterbourgon/diskv"
	"k8s.io/klog/v2"
)

var (
	// user-agent to emulate
	userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.48 Safari/537.36"

	// nonWords
	nonWordRe = regexp.MustCompile(`\W+`)

	// How long to cache by default
	DefaultMaxAge = 4 * time.Hour
)

// Request defines what can be passed in as a request
type Request struct {
	// Method type
	Method string
	// URL
	URL string
	// Referrer
	Referrer string
	// CookieJar
	Jar *cookiejar.Jar
	// Cookies
	Cookies []*http.Cookie
	// POST form values
	Form url.Values
	// Maximum age of content.
	MaxAge time.Duration
	// POST info
	ContentType string
	Body        []byte
}

// Key returns a cache-key.
func (r Request) Key() string {
	var buf bytes.Buffer

	buf.WriteString(r.Method + " ")
	buf.WriteString(r.URL + "?" + r.Form.Encode())

	for _, c := range r.Cookies {
		buf.WriteString(fmt.Sprintf("+cookie=%s", c.String()))

	}
	if r.Referrer != "" {
		buf.WriteString(fmt.Sprintf("+ref=%s", r.Referrer))
	}

	if len(r.Body) > 0 {
		buf.WriteString(fmt.Sprintf("+body=%s", r.Body))
	}

	key := nonWordRe.ReplaceAllString(buf.String(), "_")
	if len(key) > 64 {
		h := md5.New()
		_, err := io.WriteString(h, key)
		if err != nil {
			klog.Errorf("key error: %w", err)
			return fmt.Sprintf("%64.64s", key)
		}
		return fmt.Sprintf("%32.32s%x", key, h.Sum(nil))
	}
	return key
}

// Response defines which data may be cached for an HTTP response.
type Response struct {
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

type Store interface {
	Read(string) ([]byte, error)
	Write(string, []byte) error
}

// tryCache attempts a cache-only fetch.
func tryCache(req Request, cs Store) (Response, error) {
	klog.V(3).Infof("tryCache: %+v", req)

	var res Response
	cachedBytes, err := cs.Read(req.Key())
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
	if age > req.MaxAge {
		return res, fmt.Errorf("URL %s cache was too old", req.URL)
	}
	klog.V(2).Infof("Cached item: %s (cookies=%+v)", res.URL, res.Cookies)
	return res, nil
}

// applyDefault applies default request options
func applyDefaults(req Request) (Request, error) {
	// Apply defaults
	if req.MaxAge == 0 {
		req.MaxAge = DefaultMaxAge
	}
	if req.Method == "" {
		req.Method = "GET"
	}

	if req.Method == "POST" && len(req.Form) > 0 {
		req.Body = []byte(req.Form.Encode())
	}

	if req.Jar == nil {
		jar, err := cookiejar.New(nil)
		req.Jar = jar
		if err != nil {
			return req, fmt.Errorf("cookiejar: %v", err)
		}
	}

	if len(req.Cookies) == 0 {
		u, err := url.Parse(req.URL)
		if err != nil {
			return req, fmt.Errorf("url parse: %w", err)
		}
		req.Cookies = req.Jar.Cookies(u)
	}

	return req, nil
}

// Fetch wraps http.Get/http.Post behind a persistent ca
func Fetch(req Request, cs Store) (Response, error) {
	req, err := applyDefaults(req)
	if err != nil {
		return Response{}, fmt.Errorf("apply defaults: %w", err)
	}

	url := req.URL
	if req.Method == "GET" && len(req.Form) > 0 {
		url = url + "?" + req.Form.Encode()
	}

	klog.V(1).Infof("fetching %s: %+v", url, req)
	res, err := tryCache(req, cs)
	if err != nil {
		klog.V(2).Infof("MISS[%s]: %+v, tryCache returned: %v", req.Key(), req, err)
	} else {
		klog.V(2).Infof("HIT[%s]: max-age: %d", req.Key(), req.MaxAge)
		res.Cached = true
		return res, nil
	}

	client := &http.Client{Jar: req.Jar}

	getBody := bytes.NewBuffer(req.Body)

	hr, err := http.NewRequest(req.Method, url, getBody)
	if err != nil {
		return res, err
	}

	if req.Referrer != "" {
		hr.Header.Add("Referrer", req.Referrer)
	}

	hr.Header.Add("User-Agent", userAgent)

	for _, c := range req.Cookies {
		hr.AddCookie(c)
		klog.Infof("Cookie: %s", c)
	}

	if req.Method == "POST" {
		if len(req.Form) > 0 {
			hr.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		} else {
			hr.Header.Add("Content-Type", req.ContentType)
		}

	}

	cmd, err := http2curl.GetCurlCommand(hr)
	if err != nil {
		klog.Errorf("unable to convert to curl: %+v", req)
	} else {
		klog.Infof("debug: %s", cmd)
	}

	r, err := client.Do(hr)
	if err != nil {
		return res, err
	}
	klog.V(2).Infof("r: %+v", r)

	// Write the response into the cache. Mask over any failures.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return Response{}, err
	}
	cr := Response{
		URL:        req.URL,
		StatusCode: r.StatusCode,
		Header:     r.Header,
		Cookies:    req.Jar.Cookies(r.Request.URL),
		Body:       body,
		MTime:      time.Now(),
	}

	klog.Infof("Fetched %s, status=%d, cookies=%s, bytes=%d", req.URL, r.StatusCode, r.Cookies(), len(body))
	for k, v := range r.Header {
		klog.Infof("Response header: %s=%q", k, v)
	}
	for _, c := range cr.Cookies {
		klog.Infof("CookieJar: %s", c)
	}

	klog.V(2).Infof("body: %s", body)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(&cr)
	if err != nil {
		return cr, fmt.Errorf("encoding %+v: %v", cr, err)
	}
	bufBytes, err := ioutil.ReadAll(&buf)
	klog.V(4).Infof("Buf bytes: %s", bufBytes)
	if err != nil {
		klog.V(1).Infof("Failed to read back encoded response: %s", err)
	} else {
		klog.V(1).Infof("Storing %s", req.Key())
		err := cs.Write(req.Key(), bufBytes)
		if err != nil {
			klog.Errorf("unable to write %s: %v", req.Key(), err)
			return cr, nil
		}
	}
	cr.Cached = false
	return cr, nil
}

// Initialize returns an initialized cache
func Initialize() (*diskv.Diskv, error) {
	root, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("cache dir: %w", err)
	}
	cacheDir := filepath.Join(root, "campwiz")
	klog.Infof("cache dir is %s", cacheDir)
	klog.Infof("default expiry is %s", DefaultMaxAge)

	return diskv.New(diskv.Options{
		BasePath:     cacheDir,
		CacheSizeMax: 1024 * 1024 * 1024,
	}), nil
}
