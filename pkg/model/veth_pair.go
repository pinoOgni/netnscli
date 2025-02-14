package model

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// TODO: Create const file
var defaultNs = "default"

// VethPair represents a pair of veth interfaces
type VethPair struct {
	P1 Interface `yaml:"p1"`
	P2 Interface `yaml:"p2"`
}

// Interface represents a single interface of the veth couple
type Interface struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Address   string `yaml:"address" validate:"ipv4,ipv6"`
}

// setNamespace brings the interface to its p.Namespace
func (i Interface) setNamespace() error {
	if i.Namespace == defaultNs {
		return nil
	}
	log.Debugf("setting %s namespace to the %s interface", i.Namespace, i.Name)

	link, err := netlink.LinkByName(i.Name)
	if err != nil {
		return err
	}

	nsHandle, err := netns.GetFromName(i.Namespace)
	if err != nil {
		return err
	}
	defer nsHandle.Close()

	if err := netlink.LinkSetNsFd(link, int(nsHandle)); err != nil {
		return err
	}

	log.Debugf("%s is now in the %s namespace", i.Namespace, i.Namespace)
	return nil
}

// up brings up the interface
func (i Interface) up() error {
	log.Debugf("bringing up %s interface", i.Name)
	if i.Namespace != defaultNs {
		SetCurrent(i.Namespace)
	}

	link, err := netlink.LinkByName(i.Name)
	if err != nil {
		return err
	}

	if err := netlink.LinkSetUp(link); err != nil {
		return err
	}

	SetProgramNamespace()

	log.Debugf("%s interface up", i.Name)
	return nil
}

// addAddress add an ip address to the interface
func (i Interface) addAddress() error {
	log.Debugf("adding %s ip address to %s interface", i.Address, i.Name)
	if i.Namespace != defaultNs {
		SetCurrent(i.Namespace)
	}

	link, err := netlink.LinkByName(i.Name)
	if err != nil {
		return err
	}

	// TODO: The address is validated when the configuration is parsed from the yaml
	// maybe this check is useless
	if i.Address != "" {
		addr, err := netlink.ParseAddr(i.Address)
		if err != nil {
			return err
		}

		if err := netlink.AddrAdd(link, addr); err != nil {
			return err
		}
	}

	SetProgramNamespace()

	log.Debugf("ip %s added to %s interface", i.Address, i.Name)
	return nil
}

// Create instantiate the veth peer, brings up both interfaces and assigns the
// right address to each interface in case they have one
func (pair VethPair) Create() error {
	log.Debugf("creating veth pair [%s, %s]", pair.P1.Name, pair.P2.Name)
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = pair.P1.Name
	veth := &netlink.Veth{
		LinkAttrs: linkAttrs,
		PeerName:  pair.P2.Name,
	}

	err := netlink.LinkAdd(veth)
	if err != nil {
		return fmt.Errorf("failed to add veth pair: %v", err)
	}

	if err := pair.P1.setNamespace(); err != nil {
		return fmt.Errorf("failed to set namespace for veth interface %s: %v", pair.P1.Name, err)
	}

	if err := pair.P2.setNamespace(); err != nil {
		return fmt.Errorf("failed to set namespace for veth interface %s: %v", pair.P1.Name, err)
	}

	if err := pair.P1.addAddress(); err != nil {
		return fmt.Errorf("failed to set address for veth interface %s: %v", pair.P1.Name, err)
	}

	if err := pair.P2.addAddress(); err != nil {
		return fmt.Errorf("failed to set address for veth interface %s: %v", pair.P1.Name, err)
	}

	if err := pair.P1.up(); err != nil {
		return fmt.Errorf("failed to bring up veth interface %s: %v", pair.P1.Name, err)
	}

	if err := pair.P2.up(); err != nil {
		return fmt.Errorf("failed to bring up veth interface %s: %v", pair.P1.Name, err)
	}

	log.Debugf("veth pair [%s, %s] created", pair.P1.Name, pair.P2.Name)
	return nil
}

// Delete uninplemented
func (pair VethPair) Delete() error {
	panic(errors.ErrUnsupported)
}
