package app

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/internal/testutil"
)

func BenchmarkCategoryOperations(b *testing.B) {
	f := testutil.NewTestFixture(&testing.T{})
	f.CreateTestStructure()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		categories, _ := listCategories(f.TempDir)
		for _, cat := range categories {
			categoryFileCount(cat)
		}
	}
}

func BenchmarkRandomSelection(b *testing.B) {
	f := testutil.NewTestFixture(&testing.T{})
	catPath := f.CreateCategory("bench", "file1.jpg", "file2.jpg", "file3.jpg")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var stdout bytes.Buffer
		pr := &prompter{r: bufio.NewReader(strings.NewReader("s\n"))}
		runCategoryFlow(catPath, f.Cache, pr, &stdout)
	}
}

func BenchmarkCacheOperations(b *testing.B) {
	f := testutil.NewTestFixture(&testing.T{})
	catPath := f.CreateCategory("bench", "file1.jpg", "file2.jpg", "file3.jpg")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Cache.Add("file1.jpg", catPath)
		f.Cache.Load()
		f.Cache.Clear(catPath)
	}
}
