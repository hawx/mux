# mux

A request router and dispatcher.

Phases of done-ness:

1. Route based on method:

   ``` golang
   http.Handle("/items", mux.Method{
     "GET": getHandler,
     "PUT": putHandler,
   })
   ```

   Return a `405 Method Not Allowed` if no match found.

2. Default implementation of `OPTIONS` method.

3. Route based on `ContentType` header:

   ``` golang
   http.Handle("/items", mux.ContentType{
     "application/xml": xmlItemsGet,
     "application/json": jsonItemsGet,
     "*/*": itemsGet,
   })
   ```

   If no match return a `415 Unsupported Media Type`.

4. Route based on `Accept` header:

   ``` golang
   http.Handle("/items", mux.Accept{
     "application/xml": xmlItemsPut,
     "application/json": jsonItemsPut,
     "application/x-www-form-urlencoded": itemsPut,
   })
   ```

   If no match return a `406 Not Acceptable`

Current Status: 0.
