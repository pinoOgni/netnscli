package netns

import (
	"fmt"
	"os"
	"path"

	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
)

const bindMountPath = "/run/netns"

// CreaetNamespace creates a named namespace
func CreateNamespace(nsName string) error {
	ns, err := netns.NewNamed(nsName)
	defer ns.Close()
	if err != nil {
		return fmt.Errorf("cannot create namespace: %v", err)
	}
	return nil
}

// namespaceExists checks if a namespace exists
func namespaceExists(name string) (bool, error) {
	nsPath := path.Join(bindMountPath, name)
	if _, err := os.Stat(nsPath); os.IsNotExist(err) {
		return false, nil // Namespace does not exist
	} else if err != nil {
		return false, err // Other error occurred
	}
	return true, nil // Namespace exists
}

// unmountNamespace tries to unmount a namespace
func unmountNamespace(nsPath string) error {
	// Try to unmount the namespace
	if err := unix.Unmount(nsPath, unix.MNT_DETACH); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to unmount namespace %s: %w", nsPath, err)
	}
	return nil
}

// deleteNamespace deletes a namespace
func deleteNamespace(name string) error {
	nsPath := path.Join(bindMountPath, name)

	// Unmount the namespace if mounted
	if err := unmountNamespace(nsPath); err != nil {
		return fmt.Errorf("failed to unmount namespace: %w", err)
	}

	// Remove the namespace file
	if err := os.Remove(nsPath); err != nil {
		return fmt.Errorf("failed to remove namespace file %s: %w", nsPath, err)
	}

	return nil
}

// DeleteNamespace first check if a namespace exists and then delete it
func DeleteNamespace(nsName string) error {
	exists, err := namespaceExists(nsName)
	if err != nil {
		return fmt.Errorf("check namespace existence: %v", err)
	}
	if exists {
		// Attempt to delete the namespace
		err := deleteNamespace(nsName)
		if err != nil {
			return fmt.Errorf("cannot delete existing namespace: %v", err)
		}
	}
	return nil
}
