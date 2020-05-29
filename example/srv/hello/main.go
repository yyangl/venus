package main

import (
	"context"
	"fmt"
	"github.com/yyangl/venus/pb"
	"github.com/yyangl/venus/service"
)

type Hello struct {
}

func main() {
	srv := service.NewService("hello", "1", []string{"127.0.0.1:2379"})
	srv.AddHandler(service.MethodGet, "/hello", "v1", "test", func(ctx context.Context, req map[string]string) (*pb.Response, error) {
		resp := &pb.Response{}
		resp.Code = 200
		resp.Msg = "hello"
		return resp, nil
	})
	srv.AddHandler(service.MethodGet, "/he", "v1", "he", func(ctx context.Context, req map[string]string) (*pb.Response, error) {
		resp := &pb.Response{}
		resp.Code = 200
		resp.Msg = "hello"
		return resp, nil
	})
	if err := srv.Run(":0"); err != nil {
		fmt.Printf("service run error %v", err)
	}
}
