package cache

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestApplyDefaults(t *testing.T) {
	req := Request{
		URL: "/",
	}

	got, err := applyDefaults(req)
	if err != nil {
		t.Errorf("error: %v", err)
	}

	empty, err := cookiejar.New(nil)
	if err != nil {
		t.Errorf("jar: %v", err)
	}

	want := Request{
		Method:   "GET",
		URL:      "/",
		Referrer: "",
		Jar:      empty,
		MaxAge:   RecommendedMaxAge,
	}

	if diff := cmp.Diff(want, got, cmpopts.IgnoreUnexported(cookiejar.Jar{})); diff != "" {
		t.Errorf("applyDefaults() mismatch (-want +got):\n%s", diff)
	}
}

type FakeStore struct {
	seen map[string][]byte
}

func (f *FakeStore) Read(key string) ([]byte, error) {
	bs, exists := f.seen[key]
	if !exists {
		return bs, fmt.Errorf("%q not found", key)
	}
	return bs, nil
}

func (f *FakeStore) Write(key string, bs []byte) error {
	f.seen[key] = bs
	return nil
}

func TestFetch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hi")
	}))
	defer ts.Close()

	cs := &FakeStore{seen: map[string][]byte{}}
	got, err := Fetch(Request{URL: ts.URL}, cs)
	if err != nil {
		t.Errorf("fetch error: %v", err)
	}

	if got.StatusCode != 200 {
		t.Errorf("expected status code 200")
	}
	if got.Cached != false {
		t.Errorf("expected uncached result")
	}
	if string(got.Body) != "hi\n" {
		t.Errorf("got response: %q", got.Body)
	}

	// for comparison
	want := got
	want.Cached = true

	// now try with a cache
	got, err = Fetch(Request{URL: ts.URL}, cs)
	if err != nil {
		t.Errorf("fetch error: %v", err)
	}

	if diff := cmp.Diff(want, got, cmpopts.IgnoreUnexported(cookiejar.Jar{})); diff != "" {
		t.Errorf("applyDefaults() mismatch (-want +got):\n%s", diff)
	}
}
