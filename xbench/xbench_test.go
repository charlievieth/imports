package xbench

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/charlievieth/imports"
	"github.com/charlievieth/imports/gocommand"
)

func benchmarkProcess(b *testing.B, replace string) {
	name, err := filepath.Abs("../imports.go")
	if err != nil {
		b.Fatal(err)
	}
	src, err := os.ReadFile(name)
	if err != nil {
		b.Fatal(err)
	}
	if replace != "" {
		src = bytes.ReplaceAll(src, []byte(replace), []byte{})
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// TODO: is there any value to adding benchmarks where
		// the imports.Options{} are reused ???
		opts := imports.Options{
			Comments:    true,
			Fragment:    true,
			SimplifyAST: true,
			Env: &imports.ProcessEnv{
				GocmdRunner: &gocommand.Runner{},
			},
		}
		_, err := imports.Process(name, src, &opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProcess_NoChange(b *testing.B) {
	benchmarkProcess(b, "")
}

func BenchmarkProcess_MissingStdLib(b *testing.B) {
	benchmarkProcess(b, `"go/format"`)
}

func BenchmarkProcess_MissingTools(b *testing.B) {
	benchmarkProcess(b, `"golang.org/x/tools/go/ast/astutil"`)
}
