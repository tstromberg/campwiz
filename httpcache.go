// The autocamper package contains all of the brains for querying campsites.
package autocamper

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"log"
	"net/http"
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
)

// CachedHttpResponse defines which data may be cached for an HTTP response.
type CachedHttpResponse struct {
	// StatusCode is the HTTP status code value (ex: 400)
	StatusCode int
	// Body is the entire HTTP message body.
	Body []byte
	// MTime is when this value was last updated in the cache.
	MTime time.Time
}

// cachedFetch wraps http.Get/http.Post behind a persistent cache.
func cachedFetch(url string, v url.Values, maxAge time.Duration) (CachedHttpResponse, error) {
	key := []byte(url + "?" + v.Encode())
	cachedBytes, err := collection.Get(key)

	// Item is in cache, but we do not yet know if it is too old.
	if cachedBytes != nil {
		log.Printf("FOUND: %s", key)
		var cr CachedHttpResponse
		buf := bytes.NewBuffer(cachedBytes)
		dec := gob.NewDecoder(buf)
		err = dec.Decode(&cr)
		// Invalid item in cache?
		if err != nil {
			log.Printf("Failed to decode: %s", err)
		} else {
			age := time.Since(cr.MTime)
			log.Printf("Item age is %s", age)
			if age < maxAge {
				log.Printf("HIT: %s")
				return cr, nil
			} else {
				log.Printf("MISS (TOO OLD): %s", key)
			}
		}
	}

	// It's a miss.
	log.Printf("MISS: %s", key)
	// GET vs POST
	var r *http.Response
	if v != nil {
		r, err = http.Get(url)
	} else {
		r, err = http.PostForm(url, v)
	}

	// Write the response into the cache. Mask over any failures.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return CachedHttpResponse{}, err
	}
	cr := CachedHttpResponse{StatusCode: r.StatusCode, Body: body, MTime: time.Now()}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(&cr)
	if err != nil {
		log.Printf("Failed to encode response: %s", err)
	} else {
		bufBytes, err := ioutil.ReadAll(&buf)
		log.Printf("Buf bytes: %s", bufBytes)
		if err != nil {
			log.Printf("Failed to read back encoded response: %s", err)
		} else {
			log.Printf("Storing %s", key)
			collection.Set(key, bufBytes)
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
