package netlink

import (
	"fmt"

	"github.com/pinoOgni/netnscli/pkg/testbed"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const defaultNs = "default"

// CreateVethPair creates a veth pair
func CreateVethPair(vethPair testbed.VethPair) error {
	// Create the veth pair
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = vethPair.P1Name
	veth := &netlink.Veth{
		LinkAttrs: linkAttrs,
		PeerName:  vethPair.P2Name,
	}

	err := netlink.LinkAdd(veth)
	if err != nil {
		return fmt.Errorf("failed to add veth pair: %v", err)
	}

	return nil
}

// setVethPeerNs sets a peer in a network namespace
func setVethPeerNs(name, namespace string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("failed to get link %s: %v", name, err)
	}

	nsHandle, err := netns.GetFromName(namespace)
	if err != nil {
		return fmt.Errorf("failed to get namespace %s: %v", namespace, err)
	}
	defer nsHandle.Close()

	if err := netlink.LinkSetNsFd(link, int(nsHandle)); err != nil {
		return fmt.Errorf("failed to move %s to namespace %s: %v", name, namespace, err)
	}
	return nil
}

// SetVethPairNs
func SetVethPairNs(vethPair testbed.VethPair) error {

	if vethPair.P1Namespace != defaultNs {
		if err := setVethPeerNs(vethPair.P1Name, vethPair.P1Namespace); err != nil {
			return err
		}
	}

	if vethPair.P2Namespace != defaultNs {
		if err := setVethPeerNs(vethPair.P2Name, vethPair.P2Namespace); err != nil {
			return err
		}
	}
	return nil
}

// setVethPeerUp
func setVethPeerUp(name, namespace string) error {
	if namespace == defaultNs {
		link, err := netlink.LinkByName(name)
		if err != nil {
			return fmt.Errorf("failed to get link %s: %v", name, err)
		}

		if err := netlink.LinkSetUp(link); err != nil {
			return fmt.Errorf("failed to set %s up: %v", link, err)
		}
		return nil
	}

	origNS, err := netns.Get()
	if err != nil {
		return fmt.Errorf("failed to get current network namespace: %v", err)
	}
	defer origNS.Close()

	nsHandle, err := netns.GetFromName(namespace)
	if err != nil {
		return fmt.Errorf("failed to get namespace %s: %v", namespace, err)
	}
	defer nsHandle.Close()

	// Switch to the correct namespace
	if err := netns.Set(nsHandle); err != nil {
		return err
	}
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("failed to get link %s: %v", name, err)
	}

	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("failed to set %s up: %v", link, err)
	}

	// Switch back to the original namespace
	if err := netns.Set(origNS); err != nil {
		return err
	}

	return nil
}

func SetVethPairUp(vethPair testbed.VethPair) error {

	if vethPair.P1Namespace != defaultNs {
		if err := setVethPeerUp(vethPair.P1Name, vethPair.P1Namespace); err != nil {
			return err
		}
	} else {
		if err := setVethPeerUp(vethPair.P1Name, defaultNs); err != nil {
			return err
		}
	}
	if vethPair.P2Namespace != defaultNs {
		if err := setVethPeerUp(vethPair.P2Name, vethPair.P2Namespace); err != nil {
			return err
		}
	} else {
		if err := setVethPeerUp(vethPair.P2Name, defaultNs); err != nil {
			return err
		}
	}
	return nil
}

func addAddressVethPeer(peer, namespace, address string) error {
	// Save the current network namespace
	origns, _ := netns.Get()
	defer origns.Close()
	if namespace != defaultNs {
		nsHandle, err := netns.GetFromName(namespace)
		if err != nil {
			return fmt.Errorf("failed to get namespace %s: %v", namespace, err)
		}
		defer nsHandle.Close()

		netns.Set(nsHandle)
	}
	link, err := netlink.LinkByName(peer)
	if err != nil {
		return fmt.Errorf("failed to get link %s: %v", peer, err)
	}
	// Set the link up
	// TODO split
	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("failed to set %s up: %v", link, err)
	}
	if address != "" {
		addr, err := netlink.ParseAddr(address)
		if err != nil {
			return fmt.Errorf("failed to parse IP address %s: %w", address, err)
		}
		if err := netlink.AddrAdd(link, addr); err != nil {
			return fmt.Errorf("failed to add IP address %s to veth %s: %w", address, peer, err)
		}
	}
	// back in the original namespace
	if namespace != defaultNs {
		netns.Set(origns)
	}
	return nil
}

func AddAddressVethPair(vethPair testbed.VethPair) error {

	if vethPair.P1Namespace != defaultNs {
		if err := addAddressVethPeer(vethPair.P1Name, vethPair.P1Namespace, vethPair.P1IPAddress); err != nil {
			return err
		}
	} else {
		if err := addAddressVethPeer(vethPair.P1Name, defaultNs, vethPair.P1IPAddress); err != nil {
			return err
		}
	}
	if vethPair.P2Namespace != defaultNs {
		if err := addAddressVethPeer(vethPair.P2Name, vethPair.P2Namespace, vethPair.P2IPAddress); err != nil {
			return err
		}
	} else {
		if err := addAddressVethPeer(vethPair.P2Name, defaultNs, vethPair.P2IPAddress); err != nil {
			return err
		}
	}
	return nil
}

// CreateBridge creates a bridge and attach the interfaces to it
func CreateBridge(bridge testbed.Bridge) error {
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = bridge.Name
	b := &netlink.Bridge{
		LinkAttrs: linkAttrs,
	}

	if err := netlink.LinkAdd(b); err != nil {
		return fmt.Errorf("failed to add link bridge: %v", err)
	}

	return nil
}

func setUpBridge(bridge testbed.Bridge) error {
	b, err := netlink.LinkByName(bridge.Name)
	if err != nil {
		return fmt.Errorf("failed to get link %s: %v", bridge.Name, err)
	}
	// Ensure the link is of type bridge
	if b.Type() != "bridge" {
		return fmt.Errorf("link %s is not a bridge", bridge.Name)
	}
	// set the bridge up
	if err := netlink.LinkSetUp(b); err != nil {
		return fmt.Errorf("failed to set up the bridge %s: %v", bridge.Name, err)
	}
	return nil
}

func attachInterfacesToBridge(bridge testbed.Bridge) error {
	b, err := netlink.LinkByName(bridge.Name)
	if err != nil {
		return fmt.Errorf("failed to get link %s: %v", bridge.Name, err)
	}
	// Ensure the link is of type bridge
	if b.Type() != "bridge" {
		return fmt.Errorf("link %s is not a bridge", bridge.Name)
	}
	// attach the interfaces to the bridge and set them up in the default network namespace
	for _, i := range bridge.Interfaces {
		iHandle, _ := netlink.LinkByName(i)
		if err := netlink.LinkSetMaster(iHandle, b); err != nil {
			return fmt.Errorf("failed to set attach the interface %s: %v", i, err)
		}
		if err := netlink.LinkSetUp(iHandle); err != nil {
			return fmt.Errorf("failed to set up the interface %s: %v", i, err)
		}
	}
	return nil
}

func SetUpAndAttachInterfacesToBridge(bridge testbed.Bridge) error {
	if err := setUpBridge(bridge); err != nil {
		return err
	}

	if err := attachInterfacesToBridge(bridge); err != nil {
		return err
	}
	return nil
}

func DeleteBridge(bridge testbed.Bridge) error {
	link, err := netlink.LinkByName(bridge.Name)
	if err != nil {
		return fmt.Errorf("failed to get link %s: %v", bridge.Name, err)
	}
	if err := netlink.LinkDel(link); err != nil {
		return fmt.Errorf("failed to delete bridge %s: %v", bridge.Name, err)
	}
	return nil
}

// GetBridgeInterfaces gets all the interfaces attached to a bridge
func GetBridgeInterfaces(bridge testbed.Bridge) ([]netlink.Link, error) {
	// Find the bridge interface by name
	b, err := netlink.LinkByName(bridge.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to find bridge %s: %v", bridge.Name, err)
	}

	// Ensure the link is of type bridge
	if b.Type() != "bridge" {
		return nil, fmt.Errorf("link %s is not a bridge", bridge.Name)
	}

	// Get a list of all links
	allLinks, err := netlink.LinkList()
	if err != nil {
		return nil, fmt.Errorf("failed to list links: %v", err)
	}

	var bridgeLinks []netlink.Link
	// Iterate over all links and check their master
	for _, link := range allLinks {
		if link.Attrs().MasterIndex == b.Attrs().Index {
			bridgeLinks = append(bridgeLinks, link)
		}
	}
	return bridgeLinks, nil
}

// DetachInterfaceFromBridge detaches an interface from a specified bridge
func DetachInterfaceFromBridge(iface netlink.Link) error {
	// Detach the interface from the bridge
	if err := netlink.LinkSetNoMaster(iface); err != nil {
		return fmt.Errorf("failed to detach interface %s from bridge: %v", iface.Attrs().Name, err)
	}

	return nil
}

// DetachAllInterfacesFromBridge detaches all interfaces from a given bridge
func DetachAllInterfacesFromBridge(bridge testbed.Bridge) error {
	interfaces, err := GetBridgeInterfaces(bridge)
	if err != nil {
		return err
	}
	for _, iface := range interfaces {
		if err := DetachInterfaceFromBridge(iface); err != nil {
			return err
		}
	}
	return nil
}
