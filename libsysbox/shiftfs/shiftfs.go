//
// Copyright: (C) 2019 Nestybox Inc.  All rights reserved.
//

package shiftfs

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// Mark performs a shiftf mark on the given path
func Mark(path string) error {
	if err := unix.Mount(path, path, "shiftfs", 0, "mark"); err != nil {
		return fmt.Errorf("failed to mark shiftfs on %s: %v", path, err)
	}
	return nil
}

// Mount performs a shiftfs mount on the give path; the path must have a shiftfs mark on it already
func Mount(path string) error {
	if err := unix.Mount(path, path, "shiftfs", 0, ""); err != nil {
		return fmt.Errorf("failed to mount shiftfs on %s: %v", path, err)
	}
	return nil
}

func Unmount(path string) error {
	if err := unix.Unmount(path, 0); err != nil {
		return fmt.Errorf("failed to unmount %s: %v", path, err)
	}
	return nil
}