package netns

import (
	"fmt"
	"os"
	"path"

	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
)

const bindMountPath = "/run/netns"

// Add adds a network namespace
func Add(nsName string) error {
	ns, err := netns.NewNamed(nsName)
	defer func() {
		if err := ns.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}
	}()
	if err != nil {
		return fmt.Errorf("cannot add namespace: %v", err)
	}
	return nil
}

// namespaceExists checks if a network namespace exist
func namespaceExists(name string) (bool, error) {
	nsPath := path.Join(bindMountPath, name)
	if _, err := os.Stat(nsPath); os.IsNotExist(err) {
		return false, nil // network namespace does not exist
	} else if err != nil {
		return false, err // other errors
	}
	return true, nil // network namespace exists
}

// unmountNamespace tries to unmount a network namespace
func unmountNamespace(nsPath string) error {
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

// DeleteNamespace first checks if a network namespace exists and then deletes it.
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
