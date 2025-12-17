package discovery

// Instance represents a service instance
type Instance struct {
	ID       string
	Host     string
	Port     int
	Metadata map[string]string
	Weight   float64
}

// ServiceDiscovery defines the interface for service discovery
type ServiceDiscovery interface {
	// GetService returns a list of instances for the given service name
	GetService(serviceName string) ([]Instance, error)
}

// ServiceRegistry defines the interface for service registration
type ServiceRegistry interface {
	// Register registers a service instance
	Register(serviceName string, host string, port int, metadata map[string]string) error
	// Deregister deregisters a service instance
	Deregister(serviceName string, host string, port int) error
}
