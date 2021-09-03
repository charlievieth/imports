package imports

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/charlievieth/imports/gocommand"
)

func TestProjectDir(t *testing.T) {
	exp, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"", "event", "fastwalk"} {
		dir := filepath.Join(exp, s)
		pd := projectDir(dir)
		if pd != exp {
			t.Errorf("%s: got: %q want: %q", dir, pd, exp)
		}
	}

	// if the project dir is not found we should return the original
	tmp := filepath.Clean(os.TempDir())
	if pd := projectDir(tmp); pd != tmp {
		t.Errorf("%s: got: %q want: %q", tmp, pd, tmp)
	}

	if pd := projectDir("/"); pd != "/" {
		t.Errorf("%s: got: %q want: %q", "/", pd, "/")
	}
}

func TestProjectModFile(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	mod := filepath.Join(wd, "go.mod")
	fi1, err := os.Lstat(mod)
	if err != nil {
		t.Fatal(err)
	}

	got := projectModFile(wd)
	fi2, err := os.Lstat(got)
	if err != nil {
		t.Fatal(err)
	}

	if !os.SameFile(fi1, fi2) {
		t.Errorf("files not the same: got: %q want: %q", got, mod)
	}

	tmp := os.TempDir()
	if s := projectModFile(tmp); s != "" {
		t.Errorf(`exptected "" when mod file not found got: %q`, s)
	}
}

func TestGoCmdEnv(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	mod, err := filepath.Abs("./go.mod")
	if err != nil {
		t.Fatal(err)
	}

	for _, base := range []string{"", "fastwalk"} {
		p := ProcessEnv{
			GocmdRunner: &gocommand.Runner{},
			WorkingDir:  filepath.Join(wd, base),
		}
		m, err := p.goCmdEnv(nil)
		if err != nil {
			t.Fatal(err)
		}
		if m["GOMOD"] != mod {
			t.Errorf("GOMOD: got: %q want: %q", m["GOMOD"], mod)
		}
		if m["GOROOT"] != runtime.GOROOT() {
			t.Errorf("GOROOT: got: %q want: %q", m["GOROOT"], runtime.GOROOT())
		}
	}
}

func BenchmarkProjectDir(b *testing.B) {
	wd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}
	sub := []string{"testdata"}
	for c := 'A'; c <= 'G'; c++ {
		sub = append(sub, string(c))
	}
	deep := filepath.Join(wd, strings.Join(sub, string(filepath.Separator)))
	if err := os.MkdirAll(deep, 0755); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		if err := os.RemoveAll("testdata/A"); err != nil {
			b.Error(err)
		}
	})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		projectDir(deep)
	}
}

func BenchmarkGoCmdEnv_Cache(b *testing.B) {
	wd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		p := ProcessEnv{
			GocmdRunner: &gocommand.Runner{},
			WorkingDir:  wd,
		}
		if _, err := p.goCmdEnv(nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGoCmdEnv_Baseline(b *testing.B) {
	wd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		p := ProcessEnv{
			GocmdRunner: &gocommand.Runner{},
			WorkingDir:  wd,
		}
		_, err := p.invokeGo(context.TODO(), "env", append([]string{"-json"}, RequiredGoEnvVars...)...)
		if err != nil {
			b.Fatal(err)
		}
	}
}
