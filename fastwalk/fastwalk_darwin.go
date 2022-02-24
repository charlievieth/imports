//go:build darwin
// +build darwin

package fastwalk

import (
	"os"
	"syscall"
	"unsafe"
)

func readDir(dirName string, fn func(dirName, entName string, typ os.FileMode) error) error {
	fd, err := openDir(dirName)
	if err != nil {
		return &os.PathError{Op: "opendir", Path: dirName, Err: err}
	}
	defer closedir(fd)

	skipFiles := false
	var dirent syscall.Dirent
	var entptr *syscall.Dirent
	for {
		if errno := readdir_r(uintptr(fd), &dirent, &entptr); errno != 0 {
			if errno == syscall.EINTR {
				continue
			}
			return &os.PathError{Op: "readdir", Path: dirName, Err: errno}
		}
		if entptr == nil { // EOF
			break
		}
		if dirent.Ino == 0 {
			continue
		}
		typ := dtToType(dirent.Type)
		if skipFiles && typ.IsRegular() {
			continue
		}
		name := (*[len(syscall.Dirent{}.Name)]byte)(unsafe.Pointer(&dirent.Name))[:]
		for i, c := range name {
			if c == 0 {
				name = name[:i]
				break
			}
		}
		// Check for useless names before allocating a string.
		if string(name) == "." || string(name) == ".." {
			continue
		}
		nm := string(name)
		if err := fn(dirName, nm, typ); err != nil {
			if err != ErrSkipFiles {
				return err
			}
			skipFiles = true
		}
	}

	return nil
}

func dtToType(typ uint8) os.FileMode {
	switch typ {
	case syscall.DT_BLK:
		return os.ModeDevice
	case syscall.DT_CHR:
		return os.ModeDevice | os.ModeCharDevice
	case syscall.DT_DIR:
		return os.ModeDir
	case syscall.DT_FIFO:
		return os.ModeNamedPipe
	case syscall.DT_LNK:
		return os.ModeSymlink
	case syscall.DT_REG:
		return 0
	case syscall.DT_SOCK:
		return os.ModeSocket
	}
	return ^os.FileMode(0)
}

// According to https://golang.org/doc/go1.14#runtime
// A consequence of the implementation of preemption is that on Unix systems,
// including Linux and macOS systems, programs built with Go 1.14 will receive
// more signals than programs built with earlier releases.
//
// This causes opendir to sometimes fail with EINTR errors.
// We need to retry in this case.
func openDir(path string) (fd uintptr, err error) {
	for {
		fd, err := opendir(path)
		if err != syscall.EINTR {
			return fd, err
		}
	}
}
