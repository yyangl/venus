package gate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/gorilla/mux"
	"github.com/yyangl/venus/discover"
	"github.com/yyangl/venus/lb"
	"github.com/yyangl/venus/model"
	"github.com/yyangl/venus/pb"
	"github.com/yyangl/venus/utils"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	discover discover.Discover
	lbs      map[string]lb.Balance
	table    map[string]*rpcMap
}

type rpcMap struct {
	name string
	rpc  string
}

func New() (*Gate, error) {

	dis, err := discover.NewEtcdDiscover([]string{"127.0.0.1:2379"})
	if err != nil {
		return nil, err
	}

	fmt.Printf("%v", dis)

	g := &Gate{
		router: mux.NewRouter(),
		group:  make(map[string]*mux.Router),
		lbs:    make(map[string]lb.Balance),
		table:  make(map[string]*rpcMap),
	}
	g.server = &http.Server{
		Handler: g.router,
	}

	g.discover = dis
	return g, nil
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
	//g.lock.RLock()
	//defer g.lock.RUnlock()
	log.Printf("method %s prefix %s path %s", method, prefix, path)
	r := g.group[prefix]
	if r == nil {
		g.group[prefix] = g.router.PathPrefix("/" + prefix).Subrouter()
		r = g.group[prefix]
	}
	r.Handle(path, handlerFunc).Methods(method)
}

func (g *Gate) Run() error {
	log.Printf("http start port:%v", g.server.Addr)
	go g.handlerSignal()
	g.watchClient()
	g.handlerBefore()
	return g.server.ListenAndServe()
}

func (g *Gate) RunWithTLS(certFile, keyFile string) error {
	log.Printf("https start port:%v", g.server.Addr)
	go g.handlerSignal()
	g.watchClient()
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

func (g *Gate) watchClient() {
	r, err := g.discover.GetService(utils.ResolveServiceKey())
	if err != nil {
		fmt.Printf("get serve list err")
		return
	}
	fmt.Print(r)
	go func() {
		for wResp := range g.discover.Watch(utils.ResolveServiceKey()) {
			for _, ev := range wResp.Events {
				switch ev.Type {
				case mvccpb.PUT:
					servicePoint := model.ServicePoint{}
					err := json.Unmarshal(ev.Kv.Value, &servicePoint)
					if err != nil {
						log.Printf("etcd put json decode error")
					}
					log.Printf("%v", servicePoint)
					g.ResolveServiceOnline(&servicePoint)
					break
				case mvccpb.DELETE:
					k := string(ev.Kv.Key)
					idx := strings.LastIndex(k, "/")
					if idx <= 0 {
						return
					}
					id := k[idx+1:]
					k = k[:idx]
					idx = strings.LastIndex(k, "/")
					if idx <= 0 {
						return
					}
					name := k[idx+1:]
					if err != nil {
						log.Printf("etcd delete json decode error")
					}
					log.Printf("key id is %s", id)
					g.ResolveServiceOffline(name, id)
					break
				}
			}
		}
	}()
}

func (g *Gate) ResolveServiceOnline(point *model.ServicePoint) {
	g.lock.Lock()
	defer g.lock.Unlock()
	// 如果服务已经注册过
	if b, ok := g.lbs[point.Name]; ok {
		log.Printf("append new node")
		n := &lb.BlNode{Ip: point.Ip, Port: point.Port, Id: point.Id}
		b.AppendNode(n)
		return
	}
	g.lbs[point.Name] = lb.DefaultRandom
	g.lbs[point.Name].AppendNode(&lb.BlNode{Ip: point.Ip, Port: point.Port, Id: point.Id})
	log.Printf("registry new node")
	for _, node := range point.Methods {
		log.Printf("gate way append router %s%s", node.Version, node.Url)
		g.table["/"+node.Version+node.Url] = &rpcMap{
			name: point.Name,
			rpc:  node.ReqMethod,
		}
		g.AddRouter(node.Method, node.Version, node.Url, func(w http.ResponseWriter, r *http.Request) {
			node, err := g.lbs[point.Name].DoBalance()
			if err != nil {
				log.Printf("%v", err)
				return
			}
			log.Printf("%v", node)
			conn, err := grpc.Dial(fmt.Sprintf("%s:%d", node.Ip, node.Port), grpc.WithInsecure())
			if err != nil {
				log.Printf("did not connect: %v", err)
				return
			}
			defer conn.Close()

			//param := make(map[string]string)

			param := mux.Vars(r)

			rm := g.table[r.URL.Path]

			if rm == nil {
				w.Write([]byte("not found"))
				return
			}

			log.Printf("request uri %s params %v", r.URL.Path, param)

			request := &pb.Request{
				ReqData: param,
				Method:  rm.rpc,
			}

			c := pb.NewSrvClient(conn)
			log.Printf("%v", c)
			ctx := context.TODO()
			resp, err := c.Req(ctx, request)
			if err != nil {
				w.Write([]byte("rpc request err"))
			}
			d, err := json.Marshal(resp)
			if err != nil {
				w.Write([]byte("rpc request result json encode error"))
			}
			w.Write(d)
		})
		log.Printf("gate way append router %s%s over", node.Version, node.Url)
	}
}

func (g *Gate) ResolveServiceOffline(name, id string) {
	if b, ok := g.lbs[name]; ok {
		log.Printf("remove node name %s id %s", name, id)
		b.RemoveNode(id)
	}
}
