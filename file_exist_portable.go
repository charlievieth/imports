//go:build windows || plan9
// +build windows plan9

package imports

import "os"

func fileExists(name string) bool {
	_, err := os.Lstat(name)
	return err == nil
}
