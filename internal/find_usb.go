package internal

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus" // nolint:depguard
)

var logFindUSBDev = logrus.WithField("name", "FindUSBDev")

// FindUSBDev builds a map from usb device to idVendor:idProduct, e.g., /dev/ttyUSB0 -> 1a86:7523.
//
// The interests list is used to filter devices, e.g., one may pass in {"ttyUSB", "video"} and so
// only devices like /dev/ttyUSB0, /dev/ttyUSB1, /dev/video0 would be collected.
//
// CAVEAT: we don't have this method tested against many devices and/or OS releases, nor do we have
// gone through the Linux I/O subsystem specifications. Use this method carefully!
//
// nolint:funlen,gocyclo
func FindUSBDev(interests []string) map[string]string {
	// Target directories to find devices
	targets := []string{"/sys/bus/usb/devices/"}
	// Target directories evaluated from symlinks
	symlinks := make([]string, 0)
	// Device mapping
	devices := make(map[string]string)
	visited := make(map[string]bool)

	// Function to resolve a target
	resolveTarget := func(target string, evalSymlink bool) error {
		return filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// Cannot walk into the directory
				logFindUSBDev.Error(err)
				return nil
			}

			if d.IsDir() {
				// Skip directory
				return nil
			}

			if d.Type() == os.ModeSymlink {
				if !evalSymlink {
					// Skip symlink
					return nil
				}

				realpath, errEval := filepath.EvalSymlinks(path)
				if errEval != nil {
					logFindUSBDev.Errorf("cannot eval symlink %s: %v", path, err)
					return nil
				}

				// Store symlink for later resolving
				symlinks = append(symlinks, realpath)
				return nil
			}

			if d.Name() != "dev" {
				// Not a target device file
				return nil
			}

			// Found a device file
			fields := strings.Split(path, "/")
			dev := fields[len(fields)-2]

			// Skip already visited device
			if visited[dev] {
				return nil
			}
			visited[dev] = true
			logFindUSBDev.Debugf("visiting %s...", dev)

			// Check whether the device is in the interests list
			interested := false
			for _, interest := range interests {
				if strings.Contains(dev, interest) {
					interested = true
					break
				}
			}
			if !interested {
				// Not an interesting device
				return nil
			}

			// Get idVendor and idProduct corresponding to the device
			var idVB, idPB []byte
			var idV, idP string
			for offset := 0; offset < len(fields); offset++ {
				fields[len(fields)-offset-1] = "idVendor"
				idVB, err = os.ReadFile(strings.Join(fields[:len(fields)-offset], "/"))
				if err != nil {
					continue
				}

				fields[len(fields)-offset-1] = "idProduct"
				idPB, err = os.ReadFile(strings.Join(fields[:len(fields)-offset], "/"))
				if err == nil {
					idV = strings.ToLower(strings.TrimSpace(string(idVB)))
					idP = strings.ToLower(strings.TrimSpace(string(idPB)))
					break
				}
			}

			// Map dev to idV:idP
			if len(idV) > 0 && len(idP) > 0 {
				devices["/dev/"+dev] = idV + ":" + idP
			}
			return nil
		})
	}

	for _, target := range targets {
		if err := resolveTarget(target, true); err != nil {
			logFindUSBDev.Errorf("unexpected error on resolving target %s: %v", target, err)
		}
	}

	for _, target := range symlinks {
		if err := resolveTarget(target, false); err != nil {
			logFindUSBDev.Errorf("unexpected error on resolving target %s: %v", target, err)
		}
	}

	return devices
}
