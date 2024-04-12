package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"

	consulapi "github.com/hashicorp/consul/api"
)

func startHtttpServer(port int) {
	config := consulapi.DefaultConfig()
	consul, err := consulapi.NewClient(config)
	if err != nil {
		log.Println(err)
	}

	address := getHostname()
	serviceID := fmt.Sprintf("%s:%d", address, port)

	registration := &consulapi.AgentServiceRegistration{
		ID:      serviceID,
		Name:    "helloworld-server",
		Port:    port,
		Address: address,
		Check: &consulapi.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%v/ping", address, port),
			Interval:                       "2s",
			Timeout:                        "2s",
			DeregisterCriticalServiceAfter: "1s",
		},
	}

	regiErr := consul.Agent().ServiceRegister(registration)

	if regiErr != nil {
		log.Printf("Failed to register service: %s:%v ", address, port)
	} else {
		log.Printf("successfully register service: %s:%v", address, port)
	}
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	})
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, fmt.Sprintf("response from %s:%d", address, port))
	})
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 3 * time.Second, //nolint:gomnd // common
	}
	fmt.Printf("Start http server listening:%d", port)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
}
func getHostname() string {
	// hostname, _ := os.Hostname()
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}
	var hostname string
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				hostname = ipnet.IP.String()
				break
			}
		}
	}
	return hostname
}

func main() {
	min := 5000
	max := 8000
	portHttp := rand.Intn(max-min) + min
	startHtttpServer(portHttp)
}
