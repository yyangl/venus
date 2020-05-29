package service

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/yyangl/venus/model"
	"github.com/yyangl/venus/pb"
	"github.com/yyangl/venus/register"
	"github.com/yyangl/venus/utils"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	MethodGet    = http.MethodGet
	MethodPost   = http.MethodPost
	MethodPut    = http.MethodPut
	MethodDelete = http.MethodDelete
)

type HandlerFunc func(ctx context.Context, req map[string]string) (*pb.Response, error)
type RpcRequest *pb.Request
type RpcResponse *pb.Response
type Service struct {
	Server    *grpc.Server
	listener  net.Listener
	name      string
	version   string
	id        string
	addr      *net.TCPAddr
	port      string
	register  *register.Register
	endpoints []string
	methods   map[string]HandlerFunc
	rpc       []model.Methods
	mux       *sync.RWMutex
}

func (s *Service) Req(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	log.Printf("requets method %s", req.Method)
	if _, ok := s.methods[req.Method]; !ok {
		resp := &pb.Response{}
		resp.Code = 404
		resp.Msg = "Not Found Request Method"
		return resp, nil
	}
	return s.methods[req.Method](ctx, req.ReqData)
}

func NewService(name, version string, endpoints []string) *Service {
	s := grpc.NewServer()

	return &Service{
		name:      name,
		version:   version,
		id:        uuid.New().String(),
		Server:    s,
		endpoints: endpoints,
		mux:       &sync.RWMutex{},
		methods:   make(map[string]HandlerFunc),
	}
}

func (s *Service) Run(addr string) error {
	if addr == "" {
		addr = ":0"
	}

	pb.RegisterSrvServer(s.Server, s)

	t, err := net.ResolveTCPAddr("tcp", addr)
	s.addr = t
	if err != nil {
		return err
	}

	l, err := net.ListenTCP("tcp4", s.addr)

	if err != nil {
		return err
	}

	reg, err := register.NewRegister(s.endpoints, 5)
	log.Printf("service run addr %s", l.Addr().String())
	if err != nil {
		return err
	}

	s.register = reg

	port := l.Addr().(*net.TCPAddr).Port

	ip := utils.GetIP()

	uuid := uuid.New().String()

	srv := &model.ServicePoint{
		Ip:   ip,
		Port: port,
		Type: model.RpcService,
		Name: s.name,
		Id:   uuid,
	}

	if s.rpc != nil {
		srv.Methods = s.rpc
	}

	srvString, err := json.Marshal(srv)
	if err != nil {
		return errors.New("registry etcd err")
	}
	err = s.register.Put(utils.ResolveServerKey(s.name, uuid), string(srvString))

	if err != nil {
		return err
	}
	s.listener = l
	go s.handlerSignal()
	return s.Server.Serve(s.listener)
}

func (s *Service) Stop() error {
	s.Server.Stop()
	return s.register.RevokeLease()
}

func (s *Service) GetServer() *grpc.Server {
	return s.Server
}

func (s *Service) handlerSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT)
	<-ch
	log.Printf("recv system quit signal")

	if err := s.Stop(); err != nil {
		log.Printf("rpc server quit error %v", err)
	}
}

func (s *Service) AddHandler(method, url, version, rpcMethod string, handler HandlerFunc) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.methods[rpcMethod] = handler
	m := model.Methods{
		Method:    method,
		ReqMethod: rpcMethod,
		Url:       url,
		Version:   version,
	}
	s.rpc = append(s.rpc, m)
}
