package main

import (
	"github.com/yyangl/venus/gate"
	"log"
)

func main() {
	gateway := gate.New()
	//gateway.
	if err := gateway.Run(); err != nil {
		log.Fatalf("gateway run error %v", err)
	}
	//gateway.Server()
}
