package imports

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

var projectTombstonesCommon = [...]string{
	".git",
	"go.mod",
	"go.work", // go1.18
}

var projectTombstonesUncommon = [...]string{
	"glide.yaml",
	"Gopkg.toml",
	".svn",
	".hg",
}

func projectDir(dir string) string {
	dir = filepath.Clean(dir)
	orig := dir

	// fast path using comming tombstones
	for {
		for _, s := range &projectTombstonesCommon {
			if fileExists(dir + string(filepath.Separator) + s) {
				return dir
			}
		}
		d := filepath.Dir(dir)
		if d == dir {
			break
		}
		dir = d
	}

	// try with more uncommon project tombstones
	dir = orig
	for {
		for _, s := range &projectTombstonesUncommon {
			if fileExists(dir + string(filepath.Separator) + s) {
				return dir
			}
		}
		d := filepath.Dir(dir)
		if d == dir {
			break
		}
		dir = d
	}

	return orig
}

// TODO: include go.mod file in the cache key
type goEnvCacheKey struct {
	WorkDir    string
	BuildFlags string
	Env        string
	// TODO: consider hashing the contents of the mod file
	// after formatting it
	ModFile    string
	ModModTime int64
	ModSize    int64
}

func newGoEnvCacheKey(p *ProcessEnv) goEnvCacheKey {
	env := p.env()
	if len(env) > 1 {
		sort.Strings(env)
	}
	var flags []string
	if len(p.BuildFlags) != 0 {
		flags = make([]string, len(p.BuildFlags))
		copy(flags, p.BuildFlags)
		if len(flags) > 1 {
			sort.Strings(flags)
		}
	}
	dir := projectDir(p.WorkingDir)
	key := goEnvCacheKey{
		WorkDir:    dir,
		BuildFlags: strings.Join(flags, ","),
		Env:        strings.Join(env, ","),
	}
	for _, name := range []string{"go.mod", "go.work"} {
		name = dir + string(filepath.Separator) + name
		if fi, err := os.Stat(name); err == nil && fi.Mode().IsRegular() {
			key.ModFile = name
			key.ModModTime = fi.ModTime().UnixNano()
			key.ModSize = fi.Size()
		}
	}
	return key
}

type goEnvCacheEntry struct {
	once      sync.Once
	createdAt time.Time // time.Time the entry was created
	env       map[string]string
	err       error
}

func (e *goEnvCacheEntry) shouldInvalidate() bool {
	const (
		// max age of valid cache entries
		maxAge = time.Minute * 5

		// max age of invalid (error'd) cache entrie
		errInterval = time.Second * 5
	)
	d := time.Since(e.createdAt)
	return d >= maxAge || (e.err != nil && d >= errInterval)
}

func (e *goEnvCacheEntry) lazyInit(p *ProcessEnv) {
	e.once.Do(func() {
		var stdout *bytes.Buffer
		stdout, e.err = p.invokeGo(context.TODO(), "env", append([]string{"-json"}, RequiredGoEnvVars...)...)
		if e.err != nil {
			return
		}
		env := make(map[string]string, len(RequiredGoEnvVars))
		if e.err = json.Unmarshal(stdout.Bytes(), &env); e.err == nil {
			e.env = env
		}
	})
}

var goEnvCache sync.Map

func (p *ProcessEnv) goCmdEnv(_ context.Context) (map[string]string, error) {
	key := newGoEnvCacheKey(p)
	var e *goEnvCacheEntry
	v, ok := goEnvCache.Load(key)
	if !ok {
		e = &goEnvCacheEntry{createdAt: time.Now()}
		if vv, loaded := goEnvCache.LoadOrStore(key, e); loaded {
			e = vv.(*goEnvCacheEntry)
		}
	} else {
		e = v.(*goEnvCacheEntry)
	}
	e.lazyInit(p)
	if e.shouldInvalidate() {
		e = &goEnvCacheEntry{createdAt: time.Now()}
		goEnvCache.Store(key, e)
	}
	return e.env, e.err
}
