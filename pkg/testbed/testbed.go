package testbed

// VethPair represents a pair of veth interfaces
type VethPair struct {
	Name        string `mapstructure:"name"`
	P1Namespace string `mapstructure:"p1_ns"`
	P1Name      string `mapstructure:"p1_name"`
	P1IPAddress string `yaml:"p1_ip_address" mapstructure:"p1_ip_address" validate:"ipv4,ipv6"`
	P2Namespace string `mapstructure:"p2_ns"`
	P2Name      string `mapstructure:"p2_name"`
	P2IPAddress string `yaml:"p2_ip_address" mapstructure:"p2_ip_address" validate:"ipv4,ipv6"`
}

// Bridge represents a network bridge
type Bridge struct {
	Name        string   `mapstructure:"name"`
	Description string   `mapstructure:"description"`
	Interfaces  []string `mapstructure:"interfaces"`
}

type Namespace struct {
	Name        string `yaml:"name" mapstructure:"name" validate:"required"` // TODO add tag and regex validation
	Description string `yaml:"description" mapstructure:"description"`
}

type Configuration struct {
	Namespaces   []Namespace `yaml:"namespaces" mapstructure:"namespaces" validate:"required"`
	VethPairs    []VethPair  `mapstructure:"veth_pairs" validate:"-"`
	Bridges      []Bridge    `mapstructure:"bridges" validate:"-"`
	IPForwarding bool        `mapstructure:"ip_forwarding" validate:"-"`
}
