package main

// run.go — the split, end to end.

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const tableFile = "table.txt" // in scripts/split-effects: written by the split, read by --place

func split(root string, planOnly, baseline bool) error {
	serverDir := filepath.Join(root, "server")
	effDir := filepath.Join(serverDir, effectsRel)
	tbl, err := readRangeTable(filepath.Join(effDir, shardsFile))
	if err != nil {
		return err
	}
	if isSplit(effDir) {
		// Idempotent: a split tree only needs its stragglers placed.
		ms, err := check(root)
		if err != nil {
			return err
		}
		var files []string
		for _, m := range ms {
			if !strings.HasSuffix(m.path, "/") {
				files = append(files, filepath.Join(root, m.path))
			}
		}
		if len(files) > 0 {
			logf("the tree is already split; placing %d file(s)", len(files))
			if err := placeFiles(root, files); err != nil {
				return err
			}
		} else {
			logf("the tree is already split and every file is in place")
		}
		if planOnly {
			return nil
		}
		return afterSplit(root, tbl, nil)
	}
	ov, err := readOverrides(filepath.Join(toolDir(root), "overrides.txt"))
	if err != nil {
		return err
	}
	s := &splitter{root: root, serverDir: serverDir, effDir: effDir, tbl: tbl, ov: ov}
	if err := s.load(); err != nil {
		return err
	}
	s.indexPkgNames()
	s.buildUnits()
	logf("placing %d units", len(s.units))
	if err := s.place(); err != nil {
		return err
	}
	r, err := s.computeRenames()
	if err != nil {
		return err
	}
	// TestMain becomes effectstest.RunTestMain, which each test package's
	// generated TestMain calls.
	for id, o := range s.info.Defs {
		if fn, ok := o.(*types.Func); ok && id.Name == "TestMain" && isPkgLevel(fn, s.pkg) {
			r.newName[o] = "RunTestMain"
		}
	}
	s.summary(r)
	if planOnly {
		return nil
	}
	res, err := s.write(r)
	if err != nil {
		return err
	}
	logf("writing %d files (%d literals keyed, %d test paths, %d scanner literals, %d file:line keys)", len(res.files), res.keyed, s.pathRewrites, s.litRenames, s.lineRemapCount)
	for _, f := range res.files {
		if f.main && f.from != f.to {
			if err := os.MkdirAll(filepath.Dir(f.to), 0o755); err != nil {
				return err
			}
			if err := os.Rename(f.from, f.to); err != nil {
				return err
			}
		}
	}
	for _, f := range res.files {
		if err := writeFile(f.to, f.content); err != nil {
			return err
		}
	}
	if err := applyFixups(effDir); err != nil {
		return err
	}
	// Tables for --place and --check.
	if err := writeFile(filepath.Join(toolDir(root), tableFile), s.table(r)); err != nil {
		return err
	}
	var coreCards []string
	for _, u := range s.units {
		if u.idx < 0 && u.dest == keyCore && hasInitDecl(u.f.ast) {
			coreCards = append(coreCards, u.f.name)
		}
	}
	sort.Strings(coreCards)
	var cc bytes.Buffer
	cc.WriteString("# Card files (they have an init()) that stay in the core package, because\n")
	cc.WriteString("# other files use what they declare. Written by scripts/split-effects (#5906);\n")
	cc.WriteString("# never add a line: a new card goes in its shard, and a helper two cards share\n")
	cc.WriteString("# goes in a core file with no init() (docs/adding-cards.md, \"Where a card goes\").\n")
	cc.WriteString("# Removing a line, after moving that file's shared helpers out, is welcome.\n")
	for _, n := range coreCards {
		cc.WriteString(n + "\n")
	}
	if err := writeFile(filepath.Join(effDir, coreCardsFile), cc.Bytes()); err != nil {
		return err
	}
	var tokenPkg string
	for _, u := range s.units {
		if fd, ok := u.decl.(*ast.FuncDecl); ok && fd.Name.Name == "TestRealDumpEveryTokenTemplateMatchesAPrintedToken" {
			tokenPkg = pkgDir(u.dest)
		}
	}
	if err := afterSplit(root, tbl, &tokenPkg); err != nil {
		return err
	}
	if _, err := run(root, "git", "add", "-A", "server/"+effectsRel); err != nil {
		return err
	}
	if baseline {
		logf("regenerating the clone baseline (go test ./internal/cards/coverage/ -run TestNoNewExactClonesInTheCatalog -update)")
		if _, err := run(serverDir, "go", "test", "./internal/cards/coverage/", "-run", "^TestNoNewExactClonesInTheCatalog$", "-update", "-count=1"); err != nil {
			return err
		}
	}
	logf("done; run the checks in scripts/split-effects/REGENERATE.md")
	return nil
}

func hasInitDecl(f *ast.File) bool {
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Recv == nil && fd.Name.Name == "init" {
			return true
		}
	}
	return false
}

// afterSplit writes the generated files and makes the edits outside the
// effects tree. Every step is idempotent.
func afterSplit(root string, tbl *rangeTable, tokenPkg *string) error {
	serverDir := filepath.Join(root, "server")
	effDir := filepath.Join(serverDir, effectsRel)
	var testIDs []string
	for _, r := range tbl.tests {
		if _, err := os.Stat(filepath.Join(effDir, "tests", r.id)); err == nil {
			testIDs = append(testIDs, r.id)
		}
	}
	if err := writeGenerated(effDir, tbl, testIDs); err != nil {
		return err
	}
	if err := replaceOnce(filepath.Join(effDir, "catalog_complete.go"),
		"const catalogSharded = false", "const catalogSharded = true"); err != nil {
		return err
	}
	wrote, err := linkCatalog(serverDir)
	if err != nil {
		return err
	}
	logf("%d packages link the whole catalog", len(wrote))
	// CI: the effects shard runs the whole tree, with -cover again.
	if err := replaceOnce(filepath.Join(root, "scripts", "ci-test-shard.sh"),
		`"effects internal/cards/effects"`, `"effects internal/cards/effects/..."`); err != nil {
		return err
	}
	ci := filepath.Join(root, ".github", "workflows", "ci-cd.yml")
	if err := replaceOnce(ci, `        #
        # #5906: the effects shard runs without -cover. Compiling
        # internal/cards/effects with -race -cover drove the compiler to
        # ~13.9 GB and the hosted runner killed it (exit 143, no output);
        # without -cover it builds well inside the runner. Nothing reads
        # the coverage figure, which was only printed. The lasting fix is
        # splitting the package (#5906 stays open for it).
        run: |
          set -f
          cover=-cover
          if [ "${{ matrix.shard }}" = effects ]; then cover=; fi
          go test -race $cover -timeout 30m $SHARD_ARGS`, `        run: |
          set -f
          go test -race -cover -timeout 30m $SHARD_ARGS`); err != nil {
		return err
	}
	if tokenPkg != nil {
		if *tokenPkg == "" {
			return fmt.Errorf("no test package holds TestRealDumpEveryTokenTemplateMatchesAPrintedToken")
		}
		tp := "internal/cards/effects/" + *tokenPkg
		for _, e := range [][2]string{
			{`go test -c -o "$RUNNER_TEMP/effects.test" ./internal/cards/effects/`, `go test -c -o "$RUNNER_TEMP/effects.test" ./` + tp + `/`},
			{"working-directory: server/internal/cards/effects\n", "working-directory: server/" + tp + "\n"},
		} {
			if err := replaceOnce(ci, e[0], e[1]); err != nil {
				return err
			}
		}
		if err := replaceOnce(filepath.Join(root, ".github", "workflows", "e2e-nightly.yml"),
			"go test ./internal/cards/effects/ -run RealDumpEveryTokenTemplate -v",
			"go test ./"+tp+"/ -run RealDumpEveryTokenTemplate -v"); err != nil {
			return err
		}
	}
	return nil
}

func (s *splitter) summary(r *renamer) {
	count := map[string]int{}
	files := map[string]map[string]bool{}
	for _, u := range s.units {
		count[u.dest]++
		if files[u.dest] == nil {
			files[u.dest] = map[string]bool{}
		}
		files[u.dest][u.f.name] = true
	}
	var ks []string
	for k := range count {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	for _, k := range ks {
		logf("  %-14s %6d units %6d files", pkgDir(k), count[k], len(files[k]))
	}
	logf("%d names exported or renamed, %d unkeyed literals to key", len(r.newName), len(r.keyLits))
	if len(r.auto) > 0 {
		sort.Strings(r.auto)
		logf("%d collisions took an X name; review them and add lines to overrides.txt:", len(r.auto))
		for _, a := range r.auto {
			fmt.Fprintln(os.Stderr, "  "+a)
		}
	}
}

// table is table.txt: for every package-level name of the old package that
// changed package or spelling, where it went; plus renamed fields and methods.
func (s *splitter) table(r *renamer) []byte {
	var lines []string
	scope := s.pkg.Scope()
	for _, n := range scope.Names() {
		o := scope.Lookup(n)
		u := s.unitOf(o.Pos())
		if u == nil || u.testFunc {
			continue
		}
		if fn, ok := o.(*types.Func); ok && (fn.Name() == "init") {
			continue
		}
		nn := n
		if x, ok := r.newName[o]; ok {
			nn = x
		}
		if u.dest == keyCore && nn == n {
			continue
		}
		lines = append(lines, fmt.Sprintf("name %s %s %s", n, u.dest, nn))
	}
	methods := map[string]string{}
	for o, nn := range r.newName {
		switch {
		case isMethod(o):
			methods[o.Name()] = nn
		case !isPkgLevel(o, s.pkg):
			v, ok := o.(*types.Var)
			if !ok || !v.IsField() {
				continue
			}
			owner := s.fieldOwner[v]
			if owner == nil {
				continue
			}
			on := owner.Name()
			if x, ok := r.newName[owner]; ok {
				on = x
			}
			lines = append(lines, fmt.Sprintf("field %s.%s.%s %s", s.destOf(owner), on, v.Name(), nn))
		}
	}
	for m, nn := range methods {
		lines = append(lines, fmt.Sprintf("method %s %s", m, nn))
	}
	sort.Strings(lines)
	var b bytes.Buffer
	b.WriteString("# Written by scripts/split-effects (#5906). Every name of the one-package effects\n")
	b.WriteString("# that the split moved or renamed; --place reads it to rewrite files written for\n")
	b.WriteString("# the old layout. Regenerated by each split; never edit by hand.\n")
	b.WriteString("#   name <old> <package> <new>\n#   field <package>.<Type>.<old> <new>\n#   method <old> <new>\n")
	for _, l := range lines {
		b.WriteString(l + "\n")
	}
	return b.Bytes()
}
