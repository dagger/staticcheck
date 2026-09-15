package main

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestEmbedPatterns(t *testing.T) {
	data := []byte("package fixture\n//go:embed assets/*.txt all:hidden \"space file.txt\" `raw file.txt`\nvar files string\n//go:test:include ignored.txt\n//go:generate:include ignored.txt\n// go:embed ignored.txt\n")
	directives, err := goDirectivesInFile("pkg/embed.go", data)
	if err != nil {
		t.Fatal(err)
	}
	if len(directives) != 1 {
		t.Fatalf("got %d directives, want only go:embed", len(directives))
	}
	got, err := directives[0].includePatterns()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"pkg/assets/*.txt", "pkg/hidden", "pkg/space file.txt", "pkg/raw file.txt"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	_, err = (goDirective{comment: "//go:embed \"unterminated"}).includePatterns()
	if err == nil {
		t.Fatal("expected invalid quoted pattern to fail")
	}
}

func TestLocalIncludes(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"go.mod":              "module example.com/root\ngo 1.26\nreplace example.com/lib => ./lib\n",
		"root.go":             "package root\n//go:embed root.txt\nvar Root string\n",
		"lib/go.mod":          "module example.com/lib\ngo 1.26\nreplace example.com/leaf => ../leaf\n",
		"lib/lib.go":          "package lib\n//go:embed assets\nvar Assets string\n",
		"leaf/go.mod":         "module example.com/leaf\ngo 1.26\nreplace example.com/lib => ../lib\n",
		"leaf/leaf.go":        "package leaf\n//go:embed leaf.txt\nvar Leaf string\n",
		"nested/go.mod":       "module example.com/nested\ngo 1.26\n",
		"nested/nested.go":    "package nested\n//go:embed private.txt\nvar Private string\n",
		"unrelated/go.mod":    "module example.com/unrelated\ngo 1.26\n",
		"unrelated/other.go":  "package unrelated\n//go:embed secret.txt\nvar Secret string\n",
		"lib/assets/data.txt": "embedded",
	}
	for name, contents := range files {
		filename := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	index, err := indexLocal(root)
	if err != nil {
		t.Fatal(err)
	}
	includes, err := index.includesFor(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"root.txt", "**/go.mod", "**/*.cc", "lib/assets", "leaf/leaf.txt"} {
		if !slices.Contains(includes, want) {
			t.Errorf("missing %q in %v", want, includes)
		}
	}
	for _, unwanted := range []string{"nested/private.txt", "unrelated/secret.txt"} {
		if slices.Contains(includes, unwanted) {
			t.Errorf("included unrelated module asset %q", unwanted)
		}
	}
	output := t.TempDir()
	if err := run([]string{"--root", root, "--output-dir", output}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"_root_.inc", "lib.inc", "leaf.inc", "nested.inc"} {
		if _, err := os.ReadFile(filepath.Join(output, name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "lib/go.mod"), []byte("module example.com/lib\nreplace example.com/missing => ../missing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := index.includesFor("."); err == nil || !strings.Contains(err.Error(), "local replace target") {
		t.Fatalf("expected missing local replace to fail, got %v", err)
	}
}
