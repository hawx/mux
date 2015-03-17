package mux

import (
	"net/http"
)

var XmlGetItemsHandler, JsonGetItemsHandler, XmlAddItemsHandler, JsonAddItemsHandler, XmlEditItemsHandler, JsonEditItemsHandler http.Handler

func Example() {
	http.Handle("/items", Method{
		"GET": Accept{
			"application/xml":  XmlGetItemsHandler,
			"application/json": JsonGetItemsHandler,
		},
		"POST": ContentType{
			"application/xml":  XmlAddItemsHandler,
			"application/json": JsonAddItemsHandler,
		},
		"PUT": ContentType{
			"application/xml":  XmlEditItemsHandler,
			"application/json": JsonEditItemsHandler,
		},
	})
}

var GetItemsHandler, PutItemsHandler http.Handler

func ExampleMethod() {
	http.Handle("/items", Method{
		"GET": GetItemsHandler,
		"PUT": PutItemsHandler,
	})
}

var XmlItemsHandler, JsonItemsHandler http.Handler

func ExampleContentType() {
	http.Handle("/items", ContentType{
		"application/xml":  XmlItemsHandler,
		"application/json": JsonItemsHandler,
	})
}

func ExampleAccept() {
	http.Handle("/items", Accept{
		"application/xml":  XmlItemsHandler,
		"application/json": JsonItemsHandler,
	})
}
