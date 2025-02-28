package model

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

const (
	BRIDGE_TYPE = "bridge"
)

// Bridge represents a network bridge
type Bridge struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Interfaces  []string `yaml:"interfaces"`
}

var _ NetworkElement = &Bridge{}

// getFromSystem gets the bridge from the system by its name
func (b Bridge) getFromSystem() (netlink.Link, error) {
	log.Debugf("getting %s bridge from the system", b.Name)
	bridge, err := netlink.LinkByName(b.Name)
	if err != nil {
		return nil, err
	}

	// Ensure the link is of type bridge
	if bridge.Type() != BRIDGE_TYPE {
		return nil, fmt.Errorf("link %s is not a bridge", b.Name)
	}

	log.Debugf("got %s bridge from the system correctly", b.Name)
	return bridge, nil
}

// up brings up the bridge in the system
func (b Bridge) up() error {
	log.Debugf("bringing %s bridge up", b.Name)
	bridge, err := b.getFromSystem()
	if err != nil {
		return err
	}

	// set the bridge up
	if err := netlink.LinkSetUp(bridge); err != nil {
		return err
	}

	log.Debugf("%s bridge is up", b.Name)
	return nil
}

// attachInterfacesToBridge attaches all the needed interfaces to the bridge in the system
func (b Bridge) attachInterfacesToBridge() error {
	log.Debugf("attaching interfaces to %s bridge", b.Name)
	bridge, err := b.getFromSystem()
	if err != nil {
		return err
	}

	// attach the interfaces to the bridge and set them up in the default network namespace
	for _, i := range b.Interfaces {
		iHandle, _ := netlink.LinkByName(i)
		if err := netlink.LinkSetMaster(iHandle, bridge); err != nil {
			return err
		}

		if err := netlink.LinkSetUp(iHandle); err != nil {
			return err
		}
	}

	log.Debugf("attached interfaces to %s bridge", b.Name)
	return nil
}

// Create creates the bridge in the system, brings it up and attaches all the needed interfaces to it
func (b Bridge) Create() error {
	log.Debugf("creating %s bridge", b.Name)
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = b.Name
	bridge := &netlink.Bridge{
		LinkAttrs: linkAttrs,
	}

	if err := netlink.LinkAdd(bridge); err != nil {
		return fmt.Errorf("failed to add link bridge: %v", err)
	}

	if err := b.up(); err != nil {
		return fmt.Errorf("failed to bring up link bridge: %v", err)
	}

	if err := b.attachInterfacesToBridge(); err != nil {
		return fmt.Errorf("failed to attach interface to link bridge: %v", err)
	}

	log.Debugf("%s bridge created", b.Name)
	return nil
}

// detachAllInterfaces detaches all the interfaces from the bridge
func (b Bridge) detachAllInterfaces() error {
	log.Debugf("detaching all %s bridge links", b.Name)
	interfaces, err := b.getAllLinks()
	if err != nil {
		return err
	}

	for _, iface := range interfaces {
		if err := netlink.LinkSetNoMaster(iface); err != nil {
			return err
		}
	}

	log.Debugf("all %s bridge links detached", b.Name)
	return nil
}

// getAllLinks returns the list of all the links attached to the bridge, or an error
func (b Bridge) getAllLinks() ([]netlink.Link, error) {
	log.Debugf("getting all %s bridge links", b.Name)
	bridge, err := b.getFromSystem()
	if err != nil {
		return nil, err
	}

	// Get a list of all links
	allLinks, err := netlink.LinkList()
	if err != nil {
		return nil, fmt.Errorf("failed to list bridge %s links: %v", b.Name, err)
	}

	var bridgeLinks []netlink.Link
	// Iterate over all links and check their master
	for _, link := range allLinks {
		if link.Attrs().MasterIndex == bridge.Attrs().Index {
			bridgeLinks = append(bridgeLinks, link)
		}
	}

	log.Debugf("got all %s bridge links", b.Name)
	return bridgeLinks, nil
}

// Delete deletes the bridge from the system
func (b Bridge) Delete() error {
	log.Debugf("deleting %s bridge", b.Name)
	if err := b.detachAllInterfaces(); err != nil {
		return fmt.Errorf("failed to detach bridge %s interfaces: %v", b.Name, err)
	}

	link, err := b.getFromSystem()
	if err != nil {
		return fmt.Errorf("failed to get bridge %s: %v", b.Name, err)
	}

	if err := netlink.LinkDel(link); err != nil {
		return fmt.Errorf("failed to delete bridge %s: %v", b.Name, err)
	}

	log.Debugf("%s bridge deleted", b.Name)
	return nil
}
