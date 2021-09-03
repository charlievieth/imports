package imports

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

var projectTombstonesCommon = [...]string{
	".git",
	"go.mod",
}

var projectTombstonesUncommon = [...]string{
	"glide.yaml",
	"Gopkg.toml",
	".svn",
	".hg",
}

func fileExists(name string) bool {
	if runtime.GOOS != "windows" {
		var stat syscall.Stat_t
		return syscall.Lstat(name, &stat) == nil
	} else {
		_, err := os.Lstat(name)
		return err == nil
	}
}

func projectDir(dir string) string {
	dir = filepath.Clean(dir)
	orig := dir

	// fast path using comming tombstones
	for {
		for _, s := range projectTombstonesCommon {
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
		for _, s := range projectTombstonesUncommon {
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

// TODO: remove if not used
//
// TODO: use this to invalidate cache (check if a mod file was added)
func projectModFile(dir string) string {
	dir = projectDir(dir)
	if dir == "" || dir == "/" {
		return ""
	}
	p := dir + string(filepath.Separator) + "go.mod"
	if _, err := os.Lstat(p); err == nil {
		return p
	}
	return ""
}

type goEnvCacheKey struct {
	WorkDir    string
	BuildFlags string
	Env        string
}

func newGoEnvCacheKey(p *ProcessEnv) goEnvCacheKey {
	env := p.env()
	if len(env) != 0 {
		sort.Strings(env)
	}
	var flags []string
	if n := len(p.BuildFlags); n != 0 {
		flags = make([]string, n)
		copy(flags, p.BuildFlags)
		sort.Strings(flags)
	}
	return goEnvCacheKey{
		WorkDir:    projectDir(p.WorkingDir),
		BuildFlags: strings.Join(flags, ","),
		Env:        strings.Join(env, ","),
	}
}

type goEnvCacheEntry struct {
	once      sync.Once
	createdAt time.Time // time.Time the entry was created
	env       map[string]string
	err       error
}

func (e *goEnvCacheEntry) shouldInvalidate() bool {
	d := time.Since(e.createdAt)
	return d >= time.Minute || (e.err != nil && d >= time.Minute/2)
}

func (e *goEnvCacheEntry) init(p *ProcessEnv) {
	var stdout *bytes.Buffer
	stdout, e.err = p.invokeGo(context.TODO(), "env", append([]string{"-json"}, RequiredGoEnvVars...)...)
	if e.err != nil {
		return
	}
	env := make(map[string]string, len(RequiredGoEnvVars))
	if e.err = json.Unmarshal(stdout.Bytes(), &env); e.err == nil {
		e.env = env
	}
}

var goEnvCache sync.Map

func invalidateCacheEntry(key goEnvCacheKey) *goEnvCacheEntry {
	e := &goEnvCacheEntry{createdAt: time.Now()}
	goEnvCache.Store(key, e)
	return e
}

func (p *ProcessEnv) goCmdEnv(_ context.Context) (map[string]string, error) {
	key := newGoEnvCacheKey(p)
	v, ok := goEnvCache.Load(key)
	if !ok {
		v, _ = goEnvCache.LoadOrStore(key, &goEnvCacheEntry{createdAt: time.Now()})
	}
	e := v.(*goEnvCacheEntry)
	if e.shouldInvalidate() {
		e = invalidateCacheEntry(key)
	}
	e.once.Do(func() { e.init(p) })
	return e.env, e.err
}
