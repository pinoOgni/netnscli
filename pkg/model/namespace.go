package model

import (
	"fmt"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netns"
)

var (
	ProgramNamespace netns.NsHandle
)

const (
	bindMountPath = "/run/netns"
)

// Namespace represents a network namespace
type Namespace struct {
	Name        string `yaml:"name" validate:"required"` // TODO add tag and regex validation
	Description string `yaml:"description"`
}

var _ NetworkElement = &Namespace{}

// Create instantiate the new namespace in the system
func (n Namespace) Create() error {
	log.Debugf("creating %s namespace", n.Name)
	ns, err := netns.NewNamed(n.Name)
	if err != nil {
		return fmt.Errorf("cannot add namespace: %v", err)
	}

	defer func() {
		if err := ns.Close(); err != nil {
			log.Errorf("error while closing namespace file descriptor: %v", err)
		}
	}()

	SetProgramNamespace()

	log.Debugf("%s namespace created", n.Name)
	return nil
}

// Delete deletes the namespace from the system
func (n Namespace) Delete() error {
	log.Debugf("deleting %s namespace", n.Name)
	exists, err := n.exists()
	if err != nil {
		return fmt.Errorf("check namespace existence: %v", err)
	}

	if exists {
		err := netns.DeleteNamed(n.Name)
		if err != nil {
			return fmt.Errorf("cannot delete existing namespace: %v", err)
		}
	}

	log.Debugf("%s namespace created", n.Name)
	return nil
}

// exists check if the namespace does already exist in the system
func (n Namespace) exists() (bool, error) {
	log.Debugf("checking %s namespace existance", n.Name)
	nsPath := path.Join(bindMountPath, n.Name)
	if _, err := os.Stat(nsPath); os.IsNotExist(err) {
		log.Debugf("%s namespace does not exist", n.Name)
		return false, nil
	} else if err != nil {
		return false, err
	}

	log.Debugf("%s namespace exists", n.Name)
	return true, nil
}

// SetProgramNamespace sets the initial program namespace. The namespace
// that the program had when it started is the 'default' namespace where the
// 'unnamespaced' resources must reside
func SetProgramNamespace() {
	log.Debugf("setting default namespace")
	err := netns.Set(ProgramNamespace)
	if err != nil {
		log.Fatal("could not set program namespace")
	}

	log.Debugf("default namespace set")
}

// SetCurrent sets a namespace by it's name
func SetCurrent(name string) error {
	log.Debugf("setting %s namespace", name)
	nsHandle, err := netns.GetFromName(name)
	if err != nil {
		return fmt.Errorf("failed to get namespace %s: %v", name, err)
	}
	defer nsHandle.Close()

	err = netns.Set(nsHandle)
	if err != nil {
		return fmt.Errorf("failed to set %s as current namespace", name)
	}

	log.Debugf("%s namespace set", name)
	return nil
}
