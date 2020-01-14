// Package mux implements request routers.
package mux

import (
	"mime"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// Method maps http methods to different handlers. If no match is found a 405
// Method Not Allowed response is returned.
type Method map[string]http.Handler

func (route Method) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := strings.ToUpper(r.Method)

	handler, ok := route[method]
	if ok {
		handler.ServeHTTP(w, r)
		return
	}

	ks := []string{}
	for k := range route {
		ks = append(ks, k)
	}
	ks = append(ks, "OPTIONS")
	sort.Strings(ks)

	w.Header().Set("Accept", strings.Join(ks, ","))

	if method != "OPTIONS" {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// ContentType maps media-types to different handlers. Keys can contain
// wildcards, so application/* will be routed to for application/xml,
// application/json, etc. but only if there is no specific match. You can also
// define */* as a fallback handler, otherwise when no match is found a 415
// Unsupported Media Type response is returned.
type ContentType map[string]http.Handler

func (route ContentType) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = contentType
	}

	// 1. Check if we have an exact match

	handler, ok := route[mediaType]
	if ok {
		handler.ServeHTTP(w, r)
		return
	}

	possibleTypes := map[string]http.Handler{}
	for k, v := range route {
		if strings.HasSuffix(k, "/*") {
			possibleTypes[k[:len(k)-2]] = v
		}
	}

	// 2. Check if we have a handler with the same subtype

	handler, ok = possibleTypes[strings.SplitN(mediaType, "/", 2)[0]]
	if ok {
		handler.ServeHTTP(w, r)
		return
	}

	// 3. Check if we have a */* handler

	handler, ok = route["*/*"]
	if ok {
		handler.ServeHTTP(w, r)
		return
	}

	// 4. Otherwise return "415 Unsupported Media Type"

	w.WriteHeader(http.StatusUnsupportedMediaType)
}

// Accept maps Accept header values to different handlers. This will attempt to
// match the acceptable content type with the highest requested quality and
// greatest specificity. A fallback of */* can be specified which will always
// match if no others do, otherwise a 406 Not Acceptable response is returned.
type Accept map[string]http.Handler

func (route Accept) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept")

	contentTypes := parseContentTypeList(accept)

	sort.Sort(byQuality(contentTypes))

	for _, ct := range contentTypes {
		for rout, handler := range route {
			rsplit := strings.Split(rout, "/")

			// 1. Check for exact match
			// 2. Check for subtype match
			// 3. Check for wildcard
			//
			// Since the contentTypes are ordered with wildcards below specifics we
			// can check in this order with no problems.
			if ct.Type == rsplit[0] && ct.Subtype == rsplit[1] ||
				ct.Type == rsplit[0] && ct.Subtype == "*" ||
				ct.Type == "*" && ct.Subtype == "*" {

				handler.ServeHTTP(w, r)
				return
			}
		}
	}

	if handler, ok := route["*/*"]; ok {
		handler.ServeHTTP(w, r)
		return
	}

	w.WriteHeader(http.StatusNotAcceptable)
}

type byQuality []clause

func (cs byQuality) Len() int {
	return len(cs)
}

func (cs byQuality) Swap(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
}

func (cs byQuality) Less(i, j int) bool {
	return cs[i].Quality > cs[j].Quality ||
		cs[i].Type != "*" && cs[j].Type == "*" ||
		cs[i].Subtype != "*" && cs[j].Subtype == "*"
}

type clause struct {
	Type    string
	Subtype string
	Quality float32
}

func parseContentTypeList(s string) []clause {
	s = strings.Trim(s, " ")
	if len(s) == 0 {
		return []clause{}
	}

	parts := strings.Split(s, ",")
	cts := make([]clause, 0, len(parts))
	for _, part := range parts {
		ct, ok := parseContentType(part)
		if !ok {
			continue
		}

		cts = append(cts, ct)
	}

	return cts
}

func parseContentType(s string) (clause, bool) {
	mediaType, params, err := mime.ParseMediaType(s)
	if err != nil {
		return clause{}, false
	}

	q := 1.0

	if qs, ok := params["q"]; ok {
		q, err = strconv.ParseFloat(qs, 32)
		if err != nil {
			return clause{}, false
		}
	}

	mediaTypeParts := strings.Split(mediaType, "/")
	if len(mediaTypeParts) != 2 {
		return clause{}, false
	}

	return clause{mediaTypeParts[0], mediaTypeParts[1], float32(q)}, true
}
