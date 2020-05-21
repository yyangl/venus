package main

import (
	"github.com/yyangl/venus/gate"
	"log"
	"net/http"
)

var TestHandler = func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello"))
}

var Test2Handler = func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello 2"))
}

func main() {
	gateway := gate.New()
	gateway.Addr("0.0.0.0:443")
	gateway.AddAfter(func() {
		log.Printf("server stop after")
	})

	gateway.AddBefore(func() {
		log.Printf("server start before")
	})

	gateway.AddRouter("GET", "", "/test", TestHandler)
	gateway.AddRouter("GET", "/v1", "/test", Test2Handler)
	//gateway.
	if err := gateway.RunWithTLS("/Users/yyang/work/go_work/venus/example/yyang.cn.pem", "/Users/yyang/work/go_work/venus/example/yyang.cn.key.pem"); err != nil {
		log.Fatalf("gateway run error %v", err)
	}
	//gateway.Server()
}
