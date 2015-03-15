package mux

import "net/http"

type Method map[string]http.Handler

func (m Method) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, ok := m[r.Method]
	if !ok {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	handler.ServeHTTP(w, r)
}
