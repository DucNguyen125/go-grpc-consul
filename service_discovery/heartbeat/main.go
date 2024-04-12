package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/consul/api"
)

func main() {
	// Create a new Consul client
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	// Create a new service instance
	registration := new(api.AgentServiceRegistration)
	registration.ID = "example-service-1"
	registration.Name = "example-service"
	registration.Port = 8080
	registration.Tags = []string{"api"}
	check := &api.AgentServiceCheck{
		DeregisterCriticalServiceAfter: "5s",
		TTL:                            "2s",
		CheckID:                        "checkalive1",
	}
	registration.Check = check
	// Register the service with Consul
	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		log.Fatal(err)
	}

	// Send heartbeats to Consul to indicate that the service is still alive
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			err := client.Agent().PassTTL("checkalive1", "Service heartbeat")
			if err != nil {
				log.Println("Failed to send heartbeat:", err)
			}
		}
	}()

	// Define a simple HTTP handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	// Start the HTTP server
	log.Println("Starting HTTP server on port 8081...")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}
