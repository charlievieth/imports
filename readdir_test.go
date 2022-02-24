package imports

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestReadDir(t *testing.T) {
	// sameInfo := func(fi1, fi2 os.FileInfo) bool {
	// 	return fi1.Name() == fi2.Name() &&
	// 		fi1.Size() == fi2.Size() &&
	// 		fi1.Mode().Type() == fi2.Mode().Type() &&
	// 		fi1.ModTime() == fi2.ModTime() &&
	// 		fi1.IsDir() == fi2.IsDir() &&
	// 		fi1.Sys() == fi2.Sys() &&
	// 		os.SameFile(fi1, fi2)
	// }
	// _ = sameInfo
	// formatFileInfo := func(fi os.FileInfo) string {
	// 	return fmt.Sprintf("%+v", struct {
	// 		Name    string
	// 		Size    int64
	// 		Mode    os.FileMode
	// 		ModTime time.Time
	// 		IsDir   bool
	// 		Sys     string
	// 	}{
	// 		Name:    fi.Name(),
	// 		Size:    fi.Size(),
	// 		Mode:    fi.Mode().Type(),
	// 		ModTime: fi.ModTime(),
	// 		IsDir:   fi.IsDir(),
	// 		Sys:     fmt.Sprintf("%+v", fi.Sys()),
	// 	})
	// }
	// _ = formatFileInfo

	compareInfo := func(t *testing.T, fi1, fi2 os.FileInfo) {
		if fi1.Name() != fi2.Name() {
			t.Errorf("Name(%q): got: %v want: %v", fi1.Name(), fi1.Name(), fi2.Name())
		}
		if fi1.Size() != fi2.Size() {
			t.Errorf("Size(%q): got: %v want: %v", fi1.Name(), fi1.Size(), fi2.Size())
		}
		if fi1.Mode().Type() != fi2.Mode().Type() {
			t.Errorf("Mode(%q): got: %v want: %v", fi1.Name(), fi1.Mode().Type(), fi2.Mode().Type())
		}
		if fi1.ModTime() != fi2.ModTime() {
			t.Errorf("ModTime(%q): got: %v want: %v", fi1.Name(), fi1.ModTime(), fi2.ModTime())
		}
		if fi1.IsDir() != fi2.IsDir() {
			t.Errorf("IsDir(%q): got: %v want: %v", fi1.Name(), fi1.IsDir(), fi2.IsDir())
		}
		if !reflect.DeepEqual(fi1.Sys(), fi2.Sys()) {
			t.Errorf("Sys(%q): got: %#v want: %#v", fi1.Name(), fi1.Sys(), fi2.Sys())
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	want, err := ioutil.ReadDir(wd)
	if err != nil {
		t.Fatal(err)
	}
	got, err := readDir(wd)
	if err != nil {
		t.Fatal(err)
	}
	if len(want) != len(got) {
		t.Errorf("len want: %d len got: %d", len(want), len(got))
	}
	for i := range got {
		compareInfo(t, want[i], got[i])
		if t.Failed() {
			break
		}
	}
}

func BenchmarkReadDir(b *testing.B) {
	wd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		if _, err := readDir(wd); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadDir_Base(b *testing.B) {
	wd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		if _, err := ioutil.ReadDir(wd); err != nil {
			b.Fatal(err)
		}
	}
}
