package imports

import "go/build"

// Process formats and adjusts imports for the provided file.
// If opt is nil the defaults are used.
//
// Note that filename's directory influences which imports can be chosen,
// so it is important that filename be accurate.
// To process data ``as if'' it were in filename, pass the data as a non-nil src.
func Process(filename string, src []byte, opt *Options) ([]byte, error) {
	env := &fixEnv{GOPATH: build.Default.GOPATH, GOROOT: build.Default.GOROOT}
	return process(filename, src, opt, env)
}
