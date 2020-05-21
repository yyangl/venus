package gate

import (
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Gate struct {
	server   *http.Server
	router   *mux.Router
	pools    sync.Pool
	listener net.Listener
	group    map[string]*mux.Router
	lock     sync.RWMutex
	before   []func()
	after    []func()
}

func New() *Gate {
	g := &Gate{
		router: mux.NewRouter(),
		group:  make(map[string]*mux.Router, 1),
		//before: make([]func(), 1),
		//after:  make([]func(), 1),
	}
	g.server = &http.Server{
		Handler: g.router,
	}
	return g
}

func (g *Gate) Addr(addr string) {
	g.server.Addr = addr
}

func (g *Gate) AddBefore(before ...func()) {
	g.before = append(g.before, before...)
}

func (g *Gate) AddAfter(after ...func()) {
	g.after = append(g.after, after...)
}

func (g *Gate) AddRouter(method, prefix, path string, handlerFunc http.HandlerFunc) {
	g.lock.RLock()
	defer g.lock.RUnlock()
	r := g.group[prefix]
	if r == nil {
		g.group[prefix] = g.router.PathPrefix(prefix).Subrouter()
		r = g.group[prefix]
	}
	r.Handle(path, handlerFunc).Methods(method)
}

func (g *Gate) Run() error {
	log.Printf("http start port:%v", g.server.Addr)
	go g.handlerSignal()
	g.handlerBefore()
	return g.server.ListenAndServe()
}

func (g *Gate) RunWithTLS(certFile, keyFile string) error {
	log.Printf("https start port:%v", g.server.Addr)
	go g.handlerSignal()
	g.handlerBefore()
	return g.server.ListenAndServeTLS(certFile, keyFile)
}

func (g *Gate) handlerSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT)
	<-ch
	log.Printf("recv system quit signal")

	g.handlerAfter()

	if err := g.server.Shutdown(nil); err != nil {
		log.Printf("http server quit error %v", err)
	}
}

func (g *Gate) handlerAfter() {
	for _, f := range g.after {
		f()
		//log.Printf("%v", f)
	}
}
func (g *Gate) handlerBefore() {
	for _, f := range g.before {
		f()
		//log.Printf("%v", f)
	}
}
