package main

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"runtime"
	"sync/atomic"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"google.golang.org/grpc/resolver"
)

type HttpClient struct {
	addrStore []string
	next      uint32
}

func (c *HttpClient) Next() string {
	n := atomic.AddUint32(&c.next, 1)
	if len(c.addrStore) == 0 {
		return ""
	}
	return c.addrStore[(int(n)-1)%len(c.addrStore)]
}

func New() *HttpClient {
	config := consulapi.DefaultConfig()
	consul, error := consulapi.NewClient(config)
	if error != nil {
		fmt.Println(error)
	}
	services, _, _ := consul.Health().Service("helloworld-server", "", true, nil)
	addrServices := make([]string, 0, len(services))
	for _, entry := range services {
		addrServices = append(addrServices, fmt.Sprintf("%s:%d", entry.Service.Address, entry.Service.Port))
	}
	c := &HttpClient{
		addrStore: addrServices,
	}
	go c.watchServices()
	return c
}

func (c *HttpClient) watchServices() {
	config := consulapi.DefaultConfig()
	consul, error := consulapi.NewClient(config)
	if error != nil {
		fmt.Println(error)
	}
	for {
		services, _, _ := consul.Health().Service("helloworld-server", "", true, nil)
		newServices := make([]string, 0)
		for _, entry := range services {
			newServices = append(newServices, fmt.Sprintf("%s:%d", entry.Service.Address, entry.Service.Port))
		}
		if reflect.DeepEqual(c.addrStore, newServices) {
			time.Sleep(1 * time.Second)
			continue
		}
		c.addrStore = newServices
		addrs := make([]resolver.Address, len(newServices))
		for _, s := range newServices {
			addrs = append(addrs, resolver.Address{Addr: s})
		}
		time.Sleep(1 * time.Second)
	}
}

func main() {
	go func() {
		for {
			fmt.Println(runtime.NumGoroutine())
			time.Sleep(1 * time.Second)
		}
	}()
	httpClientStore := New()
	fmt.Println("--- calling http echo ---")
	for i := 0; i < 1000; i++ {
		resp, err := http.Get(fmt.Sprintf("http://%s/echo", httpClientStore.Next()))
		if err != nil {
			fmt.Println(err)
		}
		if resp != nil {
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(string(body))
			time.Sleep(500 * time.Millisecond)
		} else {
			time.Sleep(500 * time.Millisecond)
		}
	}
}
