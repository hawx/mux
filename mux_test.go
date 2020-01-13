package mux

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hawx.me/code/assert"
)

// writeHandler returns a Handler that writes the given string when called.
func writeHandler(str string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, str)
	})
}

func makeRequest(method, url string) (res *http.Response, body string, err error) {
	req, err := http.NewRequest(method, url, strings.NewReader(""))
	if err != nil {
		return
	}

	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	bodyb, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	body = string(bodyb)

	return
}

// Method

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

	res, body, err := makeRequest("PUT", ts.URL)
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

	res, body, err := makeRequest("put", ts.URL)
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

	res, body, err := makeRequest("POST", ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 405, res.StatusCode)
	assert.Equal(t, "", string(body))
	assert.Equal(t, "GET,OPTIONS,PUT", res.Header.Get("Accept"))
}

func TestMethodRoutingDefaultOptions(t *testing.T) {
	ts := httptest.NewServer(Method{
		"GET": writeHandler("GET, received"),
		"PUT": writeHandler("PUT, received"),
	})
	defer ts.Close()

	res, body, err := makeRequest("OPTIONS", ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "", string(body))
	assert.Equal(t, "GET,OPTIONS,PUT", res.Header.Get("Accept"))
}

func TestMethodRoutingCanOverrideOptions(t *testing.T) {
	ts := httptest.NewServer(Method{
		"GET":     writeHandler("GET, received"),
		"OPTIONS": writeHandler("OPTIONS, received"),
	})
	defer ts.Close()

	res, body, err := makeRequest("OPTIONS", ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "OPTIONS, received", string(body))
}

// ContentType

func makeRequestWithType(method, url, contentType string) (res *http.Response, body string, err error) {
	req, err := http.NewRequest(method, url, strings.NewReader(""))
	req.Header.Set("Content-Type", contentType)
	if err != nil {
		return
	}

	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	bodyb, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	body = string(bodyb)

	return
}

func TestContentTypeRouting(t *testing.T) {
	ts := httptest.NewServer(ContentType{
		"application/xml":  writeHandler("cool xml"),
		"application/json": writeHandler("cool json"),
	})
	defer ts.Close()

	res, body, err := makeRequestWithType("GET", ts.URL, "application/json")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "cool json", string(body))
}

func TestContentTypeRoutingParameter(t *testing.T) {
	ts := httptest.NewServer(ContentType{
		"application/xml":     writeHandler("cool xml"),
		"multipart/form-data": writeHandler("cool form"),
	})
	defer ts.Close()

	res, body, err := makeRequestWithType("GET", ts.URL, "multipart/form-data; boundary=abcdefgh")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "cool form", string(body))
}

func TestContentTypeRoutingUnknownType(t *testing.T) {
	ts := httptest.NewServer(ContentType{
		"application/xml":  writeHandler("cool xml"),
		"application/json": writeHandler("cool json"),
	})
	defer ts.Close()

	res, body, err := makeRequestWithType("GET", ts.URL, "application/dog")
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

	res, body, err := makeRequestWithType("GET", ts.URL, "application/dog")
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

	res, body, err := makeRequestWithType("GET", ts.URL, "application/dog")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "cool */*", string(body))
}

// Accept

func makeRequestWithAccept(method, url, accept string) (res *http.Response, body string, err error) {
	req, err := http.NewRequest(method, url, strings.NewReader(""))
	req.Header.Set("Accept", accept)
	if err != nil {
		return
	}

	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	bodyb, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	body = string(bodyb)

	return
}

func TestAcceptRouting(t *testing.T) {
	ts := httptest.NewServer(Accept{
		"application/xml":  writeHandler("cool xml"),
		"application/json": writeHandler("cool json"),
	})
	defer ts.Close()

	res, body, err := makeRequestWithAccept("GET", ts.URL, "application/json")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "cool json", string(body))
}

func TestAcceptRoutingWhenUnspecified(t *testing.T) {
	ts := httptest.NewServer(Accept{
		"application/xml":  writeHandler("cool xml"),
		"application/json": writeHandler("cool json"),
		"*/*":              writeHandler("cool wildcard"),
	})
	defer ts.Close()

	res, body, err := makeRequest("GET", ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "cool wildcard", string(body))
}

func TestAcceptRoutingWithList(t *testing.T) {
	ts := httptest.NewServer(Accept{
		"application/xml":  writeHandler("cool xml"),
		"application/json": writeHandler("cool json"),
	})
	defer ts.Close()

	res, body, err := makeRequestWithAccept("GET", ts.URL, "application/dog,application/json")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "cool json", string(body))
}

func TestAcceptRoutingWithWeightedList(t *testing.T) {
	ts := httptest.NewServer(Accept{
		"application/xml":  writeHandler("cool xml"),
		"application/json": writeHandler("cool json"),
	})
	defer ts.Close()

	res, body, err := makeRequestWithAccept("GET", ts.URL, "application/xml;q=0.5,application/json;q=0.8")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "cool json", string(body))
}

func TestAcceptRoutingWithWeightedListAndWildcard(t *testing.T) {
	ts := httptest.NewServer(Accept{
		"application/xml":  writeHandler("cool xml"),
		"application/json": writeHandler("cool json"),
		"image/jpeg":       writeHandler("cool jpeg"),
	})
	defer ts.Close()

	res, body, err := makeRequestWithAccept("GET", ts.URL, "application/xml;q=0.5,application/json;q=0.8,image/*;q=1.0")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "cool jpeg", string(body))
}

func TestAcceptRoutingWithUnknown(t *testing.T) {
	ts := httptest.NewServer(Accept{
		"application/xml":  writeHandler("cool xml"),
		"application/json": writeHandler("cool json"),
	})
	defer ts.Close()

	res, body, err := makeRequestWithAccept("GET", ts.URL, "application/dog")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 406, res.StatusCode)
	assert.Equal(t, "", string(body))
}

func TestAcceptRoutingWithBadMediaType(t *testing.T) {
	ts := httptest.NewServer(Accept{
		"application/xml":  writeHandler("cool xml"),
		"application/json": writeHandler("cool json"),
	})
	defer ts.Close()

	res, body, err := makeRequestWithAccept("GET", ts.URL, "whateven")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 406, res.StatusCode)
	assert.Equal(t, "", string(body))
}
