package netlink

import (
	"fmt"
	"log"

	"github.com/pinoOgni/netnscli/pkg/testbed"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const defaultNs = "default"

func createVethPair(vethPair testbed.VethPair) error {

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

func setVethNs(vethPair testbed.VethPair) error {
	fmt.Println("setVetNs ", vethPair)

	if vethPair.P1Namespace != defaultNs {
		fmt.Println("vethPair.P1Namespace != defaultNs")
		link1, err := netlink.LinkByName(vethPair.P1Name)
		if err != nil {
			log.Fatalf("Failed to get link %s: %v", vethPair.P1Name, err)
		}
		// Move the veth interfaces to the respective namespaces
		ns1Handle, err := netns.GetFromName(vethPair.P1Namespace)
		if err != nil {
			log.Fatalf("Failed to get namespace %s: %v", vethPair.P1Namespace, err)
		}
		defer ns1Handle.Close()
		// Move veth1 to ns1
		if err := netlink.LinkSetNsFd(link1, int(ns1Handle)); err != nil {
			log.Fatalf("Failed to move %s to namespace %s: %v", vethPair.P1Name, vethPair.P1Namespace, err)
		}
	}

	if vethPair.P2Namespace != defaultNs {
		fmt.Println("vethPair.P2Namespace != defaultNs")
		link2, err := netlink.LinkByName(vethPair.P2Name)
		if err != nil {
			log.Fatalf("Failed to get link %s: %v", vethPair.P2Name, err)
		}

		ns2Handle, err := netns.GetFromName(vethPair.P2Namespace)
		if err != nil {
			log.Fatalf("Failed to get namespace %s: %v", vethPair.P2Namespace, err)
		}
		defer ns2Handle.Close()

		// Move veth2 to ns2
		if err := netlink.LinkSetNsFd(link2, int(ns2Handle)); err != nil {
			log.Fatalf("Failed to move %s to namespace %s: %v", vethPair.P2Name, vethPair.P2Namespace, err)
		}
	}
	return nil
}

func CreateVethPair(v testbed.VethPair) error {

	if err := createVethPair(v); err != nil {
		return err
	}
	// set the correct namespace for a given peer
	setVethNs(v)

	return nil
}

func createBridge(bridge testbed.Bridge) error {
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = bridge.Name
	b := &netlink.Bridge{
		LinkAttrs: linkAttrs,
	}

	err := netlink.LinkAdd(b)
	if err != nil {
		return fmt.Errorf("failed to add bridge: %v", err)
	}

	// set the bridge up
	if err := netlink.LinkSetUp(b); err != nil {
		return fmt.Errorf("failed to set up the bridge %s: %v", bridge.Name, err)
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

func CreateBridge(bridge testbed.Bridge) error {
	if err := createBridge(bridge); err != nil {
		return err
	}
	return nil
}

func deleteBridge(bridge testbed.Bridge) error {
	link, err := netlink.LinkByName(bridge.Name)
	if err != nil {
		return fmt.Errorf("failed to get link %s: %v", bridge.Name, err)
	}
	if err := netlink.LinkDel(link); err != nil {
		return fmt.Errorf("failed to delete bridge %s: %v", bridge.Name, err)
	}
	return nil
}

func DeleteBridge(bridge testbed.Bridge) error {
	if err := deleteBridge(bridge); err != nil {
		return err
	}
	return nil
}

func addAddress(peer string, namespace string, address string) error {
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

func setupVethPair(vethPair testbed.VethPair) error {

	if vethPair.P1Namespace != defaultNs {
		if err := addAddress(vethPair.P1Name, vethPair.P1Namespace, vethPair.P1IPAddress); err != nil {
			return err
		}
	} else {
		if err := addAddress(vethPair.P1Name, defaultNs, vethPair.P1IPAddress); err != nil {
			return err
		}
	}
	if vethPair.P2Namespace != defaultNs {
		if err := addAddress(vethPair.P2Name, vethPair.P2Namespace, vethPair.P2IPAddress); err != nil {
			return err
		}
	} else {
		if err := addAddress(vethPair.P2Name, defaultNs, vethPair.P2IPAddress); err != nil {
			return err
		}
	}
	return nil
}

func SetupVethPairs(vethPairs []testbed.VethPair) error {
	for _, v := range vethPairs {
		if err := setupVethPair(v); err != nil {
			return err
		}
	}
	return nil
}

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
