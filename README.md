# mux

Request routers.

1. Route based on method:

   ``` golang
   http.Handle("/items", mux.Method{
     "GET": getHandler,
     "PUT": putHandler,
   })
   ```

   If no match return a `405 Method Not Allowed`. A default implementation of
   the `OPTIONS` method will return an `Allow: ...` header listing the defined
   methods.

2. Route based on `ContentType` header:

   ``` golang
   http.Handle("/items", mux.ContentType{
     "application/xml": xmlItemsGet,
     "application/json": jsonItemsGet,
     "*/*": itemsGet,
   })
   ```

   If no match return a `415 Unsupported Media Type`.

3. Route based on `Accept` header:

   ``` golang
   http.Handle("/items", mux.Accept{
     "application/xml": xmlItemsPut,
     "application/json": jsonItemsPut,
     "application/x-www-form-urlencoded": itemsPut,
   })
   ```

   If no match return a `406 Not Acceptable`.
