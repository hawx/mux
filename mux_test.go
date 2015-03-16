package mux

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// writeHandler returns a Handler that writes the given string when called.
func writeHandler(str string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, str)
	})
}

func TestOk(t *testing.T) {
	ts := httptest.NewServer(writeHandler("Hello, client\n"))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "Hello, client\n", string(greeting))
}

func TestMethodRoutingForGet(t *testing.T) {
	ts := httptest.NewServer(Method{
		"GET": writeHandler("GET, received"),
		"PUT": writeHandler("PUT, received"),
	})
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "GET, received", string(body))
}

func TestMethodRoutingForPut(t *testing.T) {
	ts := httptest.NewServer(Method{
		"GET": writeHandler("GET, received"),
		"PUT": writeHandler("PUT, received"),
	})
	defer ts.Close()

	req, err := http.NewRequest("PUT", ts.URL, strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "PUT, received", string(body))
}

func TestMethodRoutingWithNonUppercaseMethod(t *testing.T) {
	ts := httptest.NewServer(Method{
		"GET": writeHandler("GET, received"),
		"PUT": writeHandler("PUT, received"),
	})
	defer ts.Close()

	req, err := http.NewRequest("put", ts.URL, strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "PUT, received", string(body))
}

func TestMethodRoutingForMissingMethod(t *testing.T) {
	ts := httptest.NewServer(Method{
		"GET": writeHandler("GET, received"),
		"PUT": writeHandler("PUT, received"),
	})
	defer ts.Close()

	req, err := http.NewRequest("POST", ts.URL, strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 405, res.StatusCode)
	assert.Equal(t, "", string(body))
}

func TestMethodRoutingDefaultOptions(t *testing.T) {
	ts := httptest.NewServer(Method{
		"GET": writeHandler("GET, received"),
		"PUT": writeHandler("PUT, received"),
	})
	defer ts.Close()

	req, err := http.NewRequest("OPTIONS", ts.URL, strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "", string(body))
	assert.Equal(t, "GET,PUT,OPTIONS", res.Header.Get("Accept"))
}

func TestMethodRoutingCanOverrideOptions(t *testing.T) {
	ts := httptest.NewServer(Method{
		"GET":     writeHandler("GET, received"),
		"OPTIONS": writeHandler("OPTIONS, received"),
	})
	defer ts.Close()

	req, err := http.NewRequest("OPTIONS", ts.URL, strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "OPTIONS, received", string(body))
}

func TestContentTypeRouting(t *testing.T) {
	ts := httptest.NewServer(ContentType{
		"application/xml":  writeHandler("cool xml"),
		"application/json": writeHandler("cool json"),
	})
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL, strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "cool json", string(body))
}

func TestContentTypeRoutingUnknownType(t *testing.T) {
	ts := httptest.NewServer(ContentType{
		"application/xml":  writeHandler("cool xml"),
		"application/json": writeHandler("cool json"),
	})
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL, strings.NewReader(""))
	req.Header.Set("Content-Type", "application/dog")
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 415, res.StatusCode)
	assert.Equal(t, "", string(body))
}

func TestContentTypeRoutingWildcardSubtype(t *testing.T) {
	ts := httptest.NewServer(ContentType{
		"application/xml": writeHandler("cool xml"),
		"application/*":   writeHandler("cool application/*"),
	})
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL, strings.NewReader(""))
	req.Header.Set("Content-Type", "application/dog")
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "cool application/*", string(body))
}

func TestContentTypeRoutingWildcardType(t *testing.T) {
	ts := httptest.NewServer(ContentType{
		"application/xml": writeHandler("cool xml"),
		"*/*":             writeHandler("cool */*"),
	})
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL, strings.NewReader(""))
	req.Header.Set("Content-Type", "application/dog")
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "cool */*", string(body))
}
