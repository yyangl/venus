package gate

import "net/http"

type Context struct {
	responseWriter http.ResponseWriter
	request        *http.Request
	store          map[string]interface{}
}

func (c *Context) Reset(w http.ResponseWriter, r *http.Request) {
	c.responseWriter = w
	c.request = r
	c.store = make(map[string]interface{})
}