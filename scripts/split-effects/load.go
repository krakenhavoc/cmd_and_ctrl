package main

// load.go — parsing and type-checking with export data from `go list`.

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

func run(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s %s: %v\n%s", name, strings.Join(args, " "), err, errb.String())
	}
	return out.String(), nil
}

// exportData maps import paths to export-data files, filled lazily with
// `go list -export`, which compiles what it lists (from the build cache when
// it can).
type exportData struct {
	serverDir string
	mu        sync.Mutex
	files     map[string]string
}

func newExportData(serverDir string) *exportData {
	return &exportData{serverDir: serverDir, files: map[string]string{}}
}

// add lists pkgs and their dependencies with -export.
func (e *exportData) add(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"list", "-e", "-export", "-deps", "-f", "{{.ImportPath}}\t{{.Export}}\t{{if .Error}}{{.Error}}{{end}}"}, pkgs...)
	out, err := run(e.serverDir, "go", args...)
	if err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, l := range strings.Split(out, "\n") {
		p := strings.SplitN(l, "\t", 3)
		if len(p) < 2 || p[1] == "" {
			continue
		}
		e.files[p[0]] = p[1]
	}
	return nil
}

func (e *exportData) importer(fset *token.FileSet) types.Importer {
	return importer.ForCompiler(fset, "gc", func(path string) (io.ReadCloser, error) {
		e.mu.Lock()
		f, ok := e.files[path]
		e.mu.Unlock()
		if !ok {
			if err := e.add(path); err != nil {
				return nil, err
			}
			e.mu.Lock()
			f, ok = e.files[path]
			e.mu.Unlock()
			if !ok {
				return nil, fmt.Errorf("no export data for %s", path)
			}
		}
		return os.Open(f)
	})
}

// testDepsOfEffects is every dependency of the effects test variant except
// the effects tree itself. Listing the test variant of effects with -export
// would compile it, which is the build that runs out of memory (#5906).
func testDepsOfEffects(serverDir string) ([]string, error) {
	out, err := run(serverDir, "go", "list", "-deps", "-test", "-f", "{{.ImportPath}}", "./"+effectsRel+"/")
	if err != nil {
		return nil, err
	}
	var pkgs []string
	for _, l := range strings.Split(out, "\n") {
		l = strings.TrimSpace(l)
		if l == "" || strings.Contains(l, " ") || strings.HasSuffix(l, ".test") || l == effectsPath || strings.HasPrefix(l, effectsPath+"/") {
			continue
		}
		pkgs = append(pkgs, l)
	}
	sort.Strings(pkgs)
	return pkgs, nil
}

// srcFile is one parsed file of a package being rewritten.
type srcFile struct {
	name string // base name
	path string // absolute
	src  []byte
	ast  *ast.File
	tf   *token.File
}

func parseDir(fset *token.FileSet, dir string, names []string) ([]*srcFile, error) {
	// Sequential, in name order, so positions (and so every tie-break the
	// tool makes on them) are the same on every run.
	out := make([]*srcFile, 0, len(names))
	for _, n := range names {
		p := filepath.Join(dir, n)
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		f, err := parser.ParseFile(fset, p, b, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return nil, err
		}
		out = append(out, &srcFile{name: n, path: p, src: b, ast: f, tf: fset.File(f.Pos())})
	}
	return out, nil
}
