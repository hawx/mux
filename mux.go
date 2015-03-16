package mux

import (
	"net/http"
	"strings"
)

type Method map[string]http.Handler

func (m Method) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := strings.ToUpper(r.Method)

	handler, ok := m[method]
	if ok {
		handler.ServeHTTP(w, r)
		return
	}

	if !ok && method == "OPTIONS" {
		ks := []string{}
		for k, _ := range m {
			ks = append(ks, k)
		}
		ks = append(ks, "OPTIONS")

		w.Header().Set("Accept", strings.Join(ks, ","))
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

type ContentType map[string]http.Handler

// TODO: Support parameters?
func (c ContentType) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")

	// 1. Check if we have an exact match

	handler, ok := c[contentType]
	if ok {
		handler.ServeHTTP(w, r)
		return
	}

	possibleTypes := map[string]http.Handler{}
	for k, v := range c {
		if strings.HasSuffix(k, "/*") {
			possibleTypes[k[:len(k)-2]] = v
		}
	}

	// 2. Check if we have a handler with the same subtype

	handler, ok = possibleTypes[strings.SplitN(contentType, "/", 2)[0]]
	if ok {
		handler.ServeHTTP(w, r)
		return
	}

	// 3. Check if we have a */* handler

	handler, ok = c["*/*"]
	if ok {
		handler.ServeHTTP(w, r)
		return
	}

	// 4. Otherwise return "415 Unsupported Media Type"

	w.WriteHeader(http.StatusUnsupportedMediaType)
}
