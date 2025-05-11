package discovery

import (
	"common-lib/envloader"
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"log"
	"os"
	"sync"
	"time"
)

// ServiceDiscovery là interface cho service discovery
type ServiceDiscovery interface {
	// Register đăng ký service với service discovery
	Register(serviceName string, servicePort int, healthCheckURL string) error

	// Discover tìm service với tên cụ thể
	Discover(serviceName string) (string, error)

	// Deregister hủy đăng ký service
	Deregister() error
}

// ConsulServiceDiscovery triển khai ServiceDiscovery sử dụng Consul
type ConsulServiceDiscovery struct {
	client       *consul.Client
	serviceID    string
	registration *consul.AgentServiceRegistration
}

// Singleton instance
var (
	instance     ServiceDiscovery
	instanceErr  error
	instanceOnce sync.Once
)

// GetConsulServiceDiscovery trả về singleton instance của ConsulServiceDiscovery
func GetConsulServiceDiscovery() (ServiceDiscovery, error) {
	instanceOnce.Do(func() {
		instance, instanceErr = newConsulServiceDiscovery()
	})
	return instance, instanceErr
}

// newConsulServiceDiscovery tạo instance mới của ConsulServiceDiscovery (private)
func newConsulServiceDiscovery() (ServiceDiscovery, error) {
	// Lấy thông tin cấu hình từ biến môi trường sử dụng envloader
	consulHost := envloader.GetEnv("CONSUL_HOST", "localhost")
	consulPort := envloader.GetEnv("CONSUL_PORT", "8500")

	// Tạo config cho Consul client
	config := consul.DefaultConfig()
	config.Address = fmt.Sprintf("%s:%s", consulHost, consulPort)

	log.Printf("Attempting to connect to Consul at %s", config.Address)

	// Thử kết nối 3 lần với khoảng cách 2 giây
	var client *consul.Client
	var err error
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		client, err = consul.NewClient(config)
		if err == nil {
			// Test connection by querying for catalog services
			_, _, err = client.Catalog().Services(nil)
			if err == nil {
				log.Printf("Successfully connected to Consul at %s", config.Address)
				break
			}
		}

		if i < maxRetries-1 {
			log.Printf("Retry %d/%d: Could not connect to Consul: %v", i+1, maxRetries, err)
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		log.Printf("All retries failed: Could not connect to Consul: %v", err)
		return nil, err
	}

	return &ConsulServiceDiscovery{
		client: client,
	}, nil
}

// NewConsulServiceDiscovery là alias cho GetConsulServiceDiscovery để giữ khả năng tương thích
func NewConsulServiceDiscovery() (ServiceDiscovery, error) {
	return GetConsulServiceDiscovery()
}

// Register đăng ký service với Consul
func (c *ConsulServiceDiscovery) Register(serviceName string, servicePort int, healthCheckURL string) error {
	// Tạo unique service ID
	c.serviceID = fmt.Sprintf("%s-%d", serviceName, servicePort)

	// Tạo service registration
	c.registration = &consul.AgentServiceRegistration{
		ID:      c.serviceID,
		Name:    serviceName,
		Port:    servicePort,
		Address: getHostname(),
		Check: &consul.AgentServiceCheck{
			HTTP:     healthCheckURL,
			Interval: "10s",
			Timeout:  "2s",
		},
	}

	log.Printf("Registering service with Consul - ID: %s, Name: %s, Address: %s, Port: %d, Health check: %s",
		c.serviceID, serviceName, c.registration.Address, servicePort, healthCheckURL)

	// Đăng ký service với Consul với cơ chế retry
	maxRetries := 3
	var err error

	for i := 0; i < maxRetries; i++ {
		err = c.client.Agent().ServiceRegister(c.registration)
		if err == nil {
			log.Printf("Successfully registered service '%s' with Consul", serviceName)

			// Kiểm tra xem service đã thực sự được đăng ký chưa
			services, _, err := c.client.Catalog().Services(nil)
			if err != nil {
				log.Printf("Warning: Could not verify registration in catalog: %v", err)
			} else {
				log.Printf("Current services in Consul catalog: %v", getMapKeys(services))

				// Kiểm tra health check status
				checks, err := c.client.Agent().Checks()
				if err != nil {
					log.Printf("Warning: Could not verify health checks: %v", err)
				} else {
					for id, check := range checks {
						if check.ServiceID == c.serviceID {
							log.Printf("Health check for service '%s' - ID: %s, Status: %s",
								serviceName, id, check.Status)
						}
					}
				}
			}

			return nil
		}

		if i < maxRetries-1 {
			log.Printf("Retry %d/%d: Could not register service with Consul: %v", i+1, maxRetries, err)
			time.Sleep(2 * time.Second)
		}
	}

	log.Printf("All retries failed: Could not register service with Consul: %v", err)
	return err
}

// Discover tìm service bằng tên
func (c *ConsulServiceDiscovery) Discover(serviceName string) (string, error) {
	var services []*consul.ServiceEntry
	var err error
	maxRetries := 3

	// Log danh sách các services đã đăng ký
	catalogServices, _, catalogErr := c.client.Catalog().Services(nil)
	if catalogErr != nil {
		log.Printf("Warning: Could not get catalog services: %v", catalogErr)
	} else {
		log.Printf("Available services in Consul: %v", getMapKeys(catalogServices))
	}

	// Thử lấy thông tin service từ Consul nhiều lần
	for i := 0; i < maxRetries; i++ {
		// Lấy danh sách các service instances có sẵn
		services, _, err = c.client.Health().Service(serviceName, "", true, nil)
		if err == nil {
			if len(services) > 0 {
				break
			} else {
				log.Printf("Service '%s' found in catalog but no healthy instances available", serviceName)
			}
		}

		if i < maxRetries-1 {
			log.Printf("Retry %d/%d: Could not discover service '%s': %v", i+1, maxRetries, serviceName, err)
			time.Sleep(1 * time.Second)
		}
	}

	// Nếu vẫn xảy ra lỗi sau khi thử nhiều lần
	if err != nil {
		return "", fmt.Errorf("could not discover service '%s' after %d retries: %v", serviceName, maxRetries, err)
	}

	// Nếu không có service khả dụng
	if len(services) == 0 {
		return "", fmt.Errorf("no healthy instances of service '%s' found", serviceName)
	}

	// Lấy instance đầu tiên (trong thực tế có thể thêm load balancing)
	service := services[0].Service

	// Trả về URL của service
	serviceURL := fmt.Sprintf("http://%s:%d", service.Address, service.Port)
	log.Printf("Discovered service '%s' at %s", serviceName, serviceURL)
	return serviceURL, nil
}

// Helper function to get map keys as a string slice
func getMapKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Deregister hủy đăng ký service
func (c *ConsulServiceDiscovery) Deregister() error {
	if c.serviceID != "" {
		err := c.client.Agent().ServiceDeregister(c.serviceID)
		if err != nil {
			return err
		}
		log.Printf("Deregistered service '%s' from Consul", c.serviceID)
	}
	return nil
}

// getHostname lấy hostname của container hiện tại
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "localhost"
	}
	return hostname
}

// GetServiceAddress lấy địa chỉ service từ biến môi trường hoặc discovery
func GetServiceAddress(serviceName string) (string, error) {
	// Sử dụng singleton instance
	discovery, err := GetConsulServiceDiscovery()
	if err != nil {
		// Fallback: sử dụng environment variables nếu không thể kết nối với Consul
		return getServiceAddressFromEnv(serviceName), nil
	}

	// Tìm service bằng discovery
	serviceURL, err := discovery.Discover(serviceName)
	if err != nil {
		// Fallback: sử dụng environment variables nếu không tìm thấy service
		return getServiceAddressFromEnv(serviceName), nil
	}

	return serviceURL, nil
}

// getServiceAddressFromEnv lấy địa chỉ service từ biến môi trường
func getServiceAddressFromEnv(serviceName string) string {

	// Nếu không có hard code cụ thể, sử dụng biến môi trường
	var baseURL, port, protocol string

	if envloader.IsProduction() {
		baseURL = envloader.GetEnv("PROD_HOST", "")
		protocol = "https"
	} else {
		baseURL = envloader.GetEnv("DEV_HOST", "")
		protocol = "http"
	}

	// Sử dụng giá trị mặc định nếu không có biến môi trường
	if baseURL == "" {
		// Trong Docker, tên service cũng là tên host
		baseURL = serviceName
	}

	switch serviceName {
	case "users":
		port = "8081"
	default:
		port = "8080"
	}

	// Tạo service URL
	return fmt.Sprintf("%s://%s:%s", protocol, baseURL, port)
}

// BuildServiceURL constructs a complete service URL including the endpoint path
func BuildServiceURL(serviceName, endpoint string) string {
	// Use service discovery to get the service address
	//baseURL, err := GetServiceAddress(serviceName)

	baseURL := getServiceAddressFromEnv(serviceName)

	// Create the full path with service name and endpoint
	return fmt.Sprintf("%s/api/%s/%s", baseURL, serviceName, endpoint)
}
