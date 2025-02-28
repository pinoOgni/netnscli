package testbed

import (
	"fmt"
	"log"
	"os"

	"github.com/pinoOgni/netnscli/pkg/model"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Namespaces   []model.Namespace `yaml:"namespaces" validate:"required"`
	VethPairs    []model.VethPair  `yaml:"veth_pairs"`
	Bridges      []model.Bridge    `yaml:"bridges"`
	IPForwarding bool              `yaml:"ip_forwarding"`
}

func FromFile(path string) *Configuration {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Error reading configuration file: %v", err)
	}

	// Unmarshal the YAML into the Configuration struct
	var config Configuration
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error unmarshalling configuration file: %v", err)
	}

	return &config
}

func (c *Configuration) networkNodes() []model.NetworkElement {
	nodes := []model.NetworkElement{}

	for _, ns := range c.Namespaces {
		nodes = append(nodes, ns)
	}

	for _, vp := range c.VethPairs {
		nodes = append(nodes, vp)
	}

	for _, br := range c.Bridges {
		nodes = append(nodes, br)
	}

	return nodes
}

func (c *Configuration) Apply() error {
	testbedNodes := c.networkNodes()

	for in := range testbedNodes {
		node := testbedNodes[in]
		if err := node.Create(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Configuration) Delete() error {
	// delete all namespaces
	if err := c.DeleteNamespaces(); err != nil {
		return fmt.Errorf("%w: %v", fmt.Errorf("failed to delete namespaces"), err)
	}

	// delete bridges
	for _, bridge := range c.Bridges {
		if err := bridge.Delete(); err != nil {
			return fmt.Errorf("%w: %v", fmt.Errorf("failed to delete local testbed"), err)
		}
	}

	return nil
}

func (c *Configuration) DeleteNamespaces() error {
	for _, namespace := range c.Namespaces {
		if err := namespace.Delete(); err != nil {
			return fmt.Errorf("%w: %v", fmt.Errorf("failed to delete namespace"), err)
		}
	}

	return nil
}

// TODO add macvaln type
