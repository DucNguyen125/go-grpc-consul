package main

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	ecpb "google.golang.org/grpc/examples/features/proto/echo"
	"google.golang.org/grpc/resolver"
)

type exampleResolverBuilder struct{}

func (*exampleResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
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
	r := &exampleResolver{
		cc:         cc,
		addrsStore: addrServices,
	}
	addrStrs := r.addrsStore
	addrs := make([]resolver.Address, len(addrStrs))
	for _, s := range addrStrs {
		addrs = append(addrs, resolver.Address{Addr: s})
	}
	r.cc.UpdateState(resolver.State{Addresses: addrs})
	go r.watchServices()
	return r, nil
}
func (*exampleResolverBuilder) Scheme() string { return "discovery" }

type exampleResolver struct {
	cc         resolver.ClientConn
	addrsStore []string
}

func (*exampleResolver) ResolveNow(o resolver.ResolveNowOptions) {
	// do nothing
}
func (*exampleResolver) Close() {
	// do nothing
}

func (r *exampleResolver) watchServices() {
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
		if reflect.DeepEqual(r.addrsStore, newServices) {
			time.Sleep(1 * time.Second)
			continue
		}
		r.addrsStore = newServices
		addrs := make([]resolver.Address, len(newServices))
		for _, s := range newServices {
			addrs = append(addrs, resolver.Address{Addr: s})
		}
		r.cc.UpdateState(resolver.State{Addresses: addrs})
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
	resolver.Register(&exampleResolverBuilder{})

	roundrobinConn, err := grpc.Dial(
		"discovery:///server",
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Println(err)
	}
	defer roundrobinConn.Close()

	fmt.Println("--- calling helloworld.Greeter/SayHello with round_robin ---")
	hwc := ecpb.NewEchoClient(roundrobinConn)
	for i := 0; i < 1000; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := hwc.UnaryEcho(ctx, &ecpb.EchoRequest{Message: "this is examples/load_balancing"})
		if err != nil {
			fmt.Println(err)
		}
		if r != nil {
			fmt.Println(r.Message)
		}
		time.Sleep(500 * time.Millisecond)
	}
}
