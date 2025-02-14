package iface

// NetworkElement is the interface representation of a network node (veth pair, bridge, namespace, ...)
type NetworkElement interface {
	Create() error
	Delete() error
}
