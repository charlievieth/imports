//go:build !windows && !plan9
// +build !windows,!plan9

package imports

import "syscall"

func fileExists(name string) bool {
	var stat syscall.Stat_t
	return syscall.Lstat(name, &stat) == nil
}
