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
