package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratePackage(t *testing.T) {
	dir := t.TempDir()
	source := `package sample

// generate:reset
type ResetableStruct struct {
	i     int
	str   string
	strP  *string
	s     []int
	m     map[string]string
	child *ResetableStruct
}
`

	if err := os.WriteFile(filepath.Join(dir, "sample.go"), []byte(source), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	packages, err := scanPackages(dir)
	if err != nil {
		t.Fatalf("scan packages: %v", err)
	}
	if len(packages) != 1 {
		t.Fatalf("got %d packages, want 1", len(packages))
	}

	if err := generatePackage(packages[0]); err != nil {
		t.Fatalf("generate package: %v", err)
	}

	generated, err := os.ReadFile(filepath.Join(dir, generatedFileName))
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}

	got := string(generated)
	assertContains(t, got, "func (v *ResetableStruct) Reset()")
	assertContains(t, got, "resetGenZero(&v.i)")
	assertContains(t, got, "resetGenZero(&v.str)")
	assertContains(t, got, "if v.strP != nil")
	assertContains(t, got, "resetGenZero(v.strP)")
	assertContains(t, got, "v.s = v.s[:0]")
	assertContains(t, got, "clear(v.m)")
	assertContains(t, got, "if v.child != nil")
	assertContains(t, got, ".Reset()")

	assertPackageTypeChecks(t, dir)
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()

	if !strings.Contains(s, substr) {
		t.Fatalf("generated source does not contain %q:\n%s", substr, s)
	}
}

func assertPackageTypeChecks(t *testing.T, dir string) {
	t.Helper()

	fset := token.NewFileSet()
	files := make([]*ast.File, 0, 2)

	for _, name := range []string{"sample.go", generatedFileName} {
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		files = append(files, file)
	}

	conf := types.Config{GoVersion: "go1.25"}
	if _, err := conf.Check("sample", fset, files, nil); err != nil {
		t.Fatalf("type-check generated package: %v", err)
	}
}
