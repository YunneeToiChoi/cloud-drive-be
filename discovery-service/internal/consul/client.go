package consul

import (
	"fmt"
	"log"

	consulapi "github.com/hashicorp/consul/api"
)

// ServiceRegistry is an interface for service registration and discovery
type ServiceRegistry interface {
	Register(name, address string, port int, tags []string) error
	Deregister(serviceID string) error
	GetService(name string) ([]*consulapi.ServiceEntry, error)
}

// ConsulClient implements ServiceRegistry using Consul
type ConsulClient struct {
	client *consulapi.Client
}

// NewConsulClient creates a new Consul client
func NewConsulClient(address string) (*ConsulClient, error) {
	config := consulapi.DefaultConfig()
	config.Address = address

	client, err := consulapi.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &ConsulClient{
		client: client,
	}, nil
}

// Register registers a service with Consul
func (c *ConsulClient) Register(name, address string, port int, tags []string) error {
	serviceID := fmt.Sprintf("%s-%s-%d", name, address, port)

	reg := &consulapi.AgentServiceRegistration{
		ID:      serviceID,
		Name:    name,
		Address: address,
		Port:    port,
		Tags:    tags,
		Check: &consulapi.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", address, port),
			Interval:                       "10s",
			Timeout:                        "1s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	err := c.client.Agent().ServiceRegister(reg)
	if err != nil {
		return err
	}

	log.Printf("Registered service %s with Consul", serviceID)
	return nil
}

// Deregister deregisters a service from Consul
func (c *ConsulClient) Deregister(serviceID string) error {
	err := c.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		return err
	}

	log.Printf("Deregistered service %s from Consul", serviceID)
	return nil
}

// GetService gets a service from Consul
func (c *ConsulClient) GetService(name string) ([]*consulapi.ServiceEntry, error) {
	services, _, err := c.client.Health().Service(name, "", true, nil)
	if err != nil {
		return nil, err
	}

	return services, nil
}
