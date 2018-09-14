package memdb

import (
	"path/filepath"
	"runtime"
	"testing"
)

func init() {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(file), "../../transformer/")
	abs, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}
	Load(abs)
}

func TestSearch(t *testing.T) {
	result := Search("a")
	if len(result) > 10 {
		t.Fail()
	}
}

func BenchmarkSearch(b *testing.B) {
	b.ResetTimer()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		x := Search("a")
		if len(x) > 10 {
			b.FailNow()
		}
	}
}
