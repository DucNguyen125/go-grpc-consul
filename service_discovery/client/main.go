package main

import (
	"fmt"
	"reflect"
	"runtime"
	"time"

	consulapi "github.com/hashicorp/consul/api"
)

func main() {
	go func() {
		for {
			fmt.Println(runtime.NumGoroutine())
			time.Sleep(1 * time.Second)
		}
	}()
	lastChange := []*consulapi.ServiceEntry{}
	for {
		config := consulapi.DefaultConfig()
		consul, error := consulapi.NewClient(config)
		if error != nil {
			fmt.Println(error)
		}
		services, _, _ := consul.Health().Service("helloworld-server", "", true, nil)
		if reflect.DeepEqual(lastChange, services) {
			time.Sleep(1 * time.Second)
			continue
		}
		lastChange = services
		fmt.Println(len(services))
		for _, entry := range services {
			fmt.Println(entry.Service.Address, entry.Service.Port)
		}
		time.Sleep(1 * time.Second)
	}
}
