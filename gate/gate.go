package gate

import (
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"sync"
)

type Gate struct {
	server   *http.Server
	router   *mux.Router
	pools    sync.Pool
	listener net.Listener
	group    map[string]*mux.Router
	lock     sync.RWMutex
	addr     string
	tls      bool
}

func New() *Gate {
	g := &Gate{
		router: mux.NewRouter(),
		group:  make(map[string]*mux.Router, 1),
		tls:    false,
		addr:   "",
	}
	//g.pools.New = func() interface{} {
	//	return NewContext(nil, nil)
	//}
	return g
}

func (g *Gate) addRouter(method, prefix, path string, handler http.Handler) {
	g.lock.RLock()
	defer g.lock.RUnlock()
	r := g.group[prefix]
	if r == nil {
		g.group[prefix] = g.router.PathPrefix(prefix).Subrouter()
		r = g.group[prefix]
	}
	r.Handle(path, handler).Methods(method)
}

func (g *Gate) Run() error {
	return g.server.ListenAndServe()
}
