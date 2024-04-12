package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"

	consulapi "github.com/hashicorp/consul/api"
)

func main() {
	port := serviceRegistryWithConsul()
	log.Println("Starting Hello World Server...")
	http.HandleFunc("/helloworld", helloworld)
	http.HandleFunc("/check", check)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func serviceRegistryWithConsul() int {
	config := consulapi.DefaultConfig()
	consul, err := consulapi.NewClient(config)
	if err != nil {
		log.Println(err)
	}

	port := getPort()
	address := getHostname()
	serviceID := fmt.Sprintf("%s:%d", address, port)

	registration := &consulapi.AgentServiceRegistration{
		ID:      serviceID,
		Name:    "helloworld-server",
		Port:    port,
		Address: address,
		Check: &consulapi.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%v/check", address, port),
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
	return port
}

func helloworld(w http.ResponseWriter, r *http.Request) {
	log.Println("helloworld service is called.")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello world.")
}

func check(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Consul check")
}

func getPort() int {
	min := 5000
	max := 8000
	port := rand.Intn(max-min) + min
	return port
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
