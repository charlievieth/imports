package imports

import (
	"io/fs"
	"os"
	"sync"
	"time"
)

type fileInfo struct {
	once sync.Once
	ent  fs.DirEntry
	info fs.FileInfo
}

func (f *fileInfo) Name() string      { return f.ent.Name() }
func (f *fileInfo) Mode() fs.FileMode { return f.ent.Type() }
func (f *fileInfo) IsDir() bool       { return f.ent.IsDir() }

// The Size, ModTime, and Sys are not used by go/build but implement
// them anyway (in case that changes) and make access thread-safe.

func (f *fileInfo) initInfo() {
	f.info, _ = f.ent.Info()
}

func (f *fileInfo) Size() int64 {
	f.once.Do(f.initInfo)
	if f.info != nil {
		return f.info.Size()
	}
	return 0
}

func (f *fileInfo) ModTime() time.Time {
	f.once.Do(f.initInfo)
	if f.info != nil {
		return f.info.ModTime()
	}
	return time.Time{}
}

func (f *fileInfo) Sys() interface{} {
	f.once.Do(f.initInfo)
	if f.info != nil {
		return f.info.Sys()
	}
	return nil
}

// readDir is a faster version of ioutil.ReadDir that uses os.ReadDir
// and returns a wrapper around fs.FileInfo.
//
// This is roughly 3.5-4x faster than ioutil.ReadDir and is used heavily
// by the build.Context when importing packages.
func readDir(dirname string) ([]fs.FileInfo, error) {
	var fis []fs.FileInfo
	ents, err := os.ReadDir(dirname)
	if len(ents) != 0 {
		fis = make([]fs.FileInfo, len(ents))
		for i, d := range ents {
			fis[i] = &fileInfo{ent: d}
		}
	}
	return fis, err
}
