package main

// layout.go — where a file of the effects tree belongs, and --check.
//
// The layout after the split (#5906):
//
//	internal/cards/effects/               package effects      core: Spec, Register, every seam and
//	                                                           shared helper, and the card files listed in
//	                                                           core_cards.txt (their declarations are other
//	                                                           files' helpers)
//	internal/cards/effects/cards/         package cards        the aggregator: blank-imports every shard
//	internal/cards/effects/cards/<id>/    package cards<id>    card files only, one init() each
//	internal/cards/effects/effectstest/   package effectstest  test helpers lifted out of _test.go files
//	internal/cards/effects/tests/<id>/    package tests<id>    _test.go files only, plus a generated main_test.go
//
// A card file's shard and a test file's package are chosen by its name, from
// the frozen range table shards.txt. Rebalancing is a deliberate edit of that
// table, never a side effect of growth.

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	modPath     = "github.com/krakenhavoc/cmd_and_ctrl/server"
	effectsRel  = "internal/cards/effects" // relative to server/
	effectsPath = modPath + "/" + effectsRel

	shardsFile    = "shards.txt"     // in the effects directory: the frozen range table
	coreCardsFile = "core_cards.txt" // in the effects directory: card files allowed to stay in core

	// The fix-it command every message names, run from the repository root.
	placeCmd = "scripts/split-effects.sh --place"
)

// Package keys used throughout the tool.
const (
	keyCore = "core"
	keyTst  = "effectstest"
)

func shardKey(id string) string { return "c/" + id }
func testKey(id string) string  { return "t/" + id }
func isShardKey(k string) bool  { return strings.HasPrefix(k, "c/") }
func isTestKey(k string) bool   { return strings.HasPrefix(k, "t/") }

// pkgDir is the directory of a package key, relative to the effects directory.
func pkgDir(k string) string {
	switch {
	case k == keyCore:
		return "."
	case k == keyTst:
		return "effectstest"
	case isShardKey(k):
		return "cards/" + k[2:]
	case isTestKey(k):
		return "tests/" + k[2:]
	}
	panic("pkgDir: " + k)
}

// pkgName is the Go package name of a package key.
func pkgName(k string) string {
	switch {
	case k == keyCore:
		return "effects"
	case k == keyTst:
		return "effectstest"
	case isShardKey(k):
		return "cards" + k[2:]
	case isTestKey(k):
		return "tests" + k[2:]
	}
	panic("pkgName: " + k)
}

func importPath(k string) string {
	d := pkgDir(k)
	if d == "." {
		return effectsPath
	}
	return effectsPath + "/" + d
}

// keyOfImportPath is the inverse of importPath ("" if it is not one of ours).
func keyOfImportPath(p string) string {
	switch {
	case p == effectsPath:
		return keyCore
	case p == effectsPath+"/effectstest":
		return keyTst
	case strings.HasPrefix(p, effectsPath+"/cards/"):
		return shardKey(strings.TrimPrefix(p, effectsPath+"/cards/"))
	case strings.HasPrefix(p, effectsPath+"/tests/"):
		return testKey(strings.TrimPrefix(p, effectsPath+"/tests/"))
	}
	return ""
}

type shardRow struct {
	id, from string
}

// rangeTable is shards.txt: for each kind, rows sorted by their lower bound.
// A file name (lower-cased) belongs to the last row whose bound is <= it.
type rangeTable struct {
	cards, tests []shardRow
}

var shardID = regexp.MustCompile(`^[a-z][a-z0-9]*$`)

func readRangeTable(path string) (*rangeTable, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	t := &rangeTable{}
	sc := bufio.NewScanner(f)
	n := 0
	for sc.Scan() {
		n++
		line := sc.Text()
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fs := strings.Fields(line)
		if len(fs) != 3 {
			return nil, fmt.Errorf("%s:%d: want <card|test> <id> <from>, got %q", path, n, line)
		}
		if !shardID.MatchString(fs[1]) {
			return nil, fmt.Errorf("%s:%d: shard id %q is not lower-case letters and digits", path, n, fs[1])
		}
		row := shardRow{fs[1], fs[2]}
		switch fs[0] {
		case "card":
			t.cards = append(t.cards, row)
		case "test":
			t.tests = append(t.tests, row)
		default:
			return nil, fmt.Errorf("%s:%d: kind %q is neither card nor test", path, n, fs[0])
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	for kind, rows := range map[string][]shardRow{"card": t.cards, "test": t.tests} {
		if len(rows) == 0 {
			return nil, fmt.Errorf("%s: no %s rows", path, kind)
		}
		if rows[0].from != "0" {
			return nil, fmt.Errorf("%s: the first %s row must start at 0, got %q", path, kind, rows[0].from)
		}
		seen := map[string]bool{}
		for i, r := range rows {
			if seen[r.id] {
				return nil, fmt.Errorf("%s: %s shard %q is listed twice", path, kind, r.id)
			}
			seen[r.id] = true
			if i > 0 && rows[i-1].from >= r.from {
				return nil, fmt.Errorf("%s: %s rows must be sorted by their bound: %q then %q", path, kind, rows[i-1].from, r.from)
			}
		}
	}
	return t, nil
}

func pick(rows []shardRow, name string) string {
	name = strings.ToLower(filepath.Base(name))
	i := sort.Search(len(rows), func(i int) bool { return rows[i].from > name }) - 1
	if i < 0 {
		i = 0
	}
	return rows[i].id
}

func (t *rangeTable) cardShard(name string) string { return shardKey(pick(t.cards, name)) }
func (t *rangeTable) testPkg(name string) string   { return testKey(pick(t.tests, name)) }

// hasInit reports whether a Go source declares a top-level func init():
// the mark of a card file (it registers itself).
func hasInit(path string) (bool, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		return false, err
	}
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Recv == nil && fd.Name.Name == "init" {
			return true, nil
		}
	}
	return false, nil
}

func readLines(path string) (map[string]bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, l := range strings.Split(string(b), "\n") {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		out[l] = true
	}
	return out, nil
}

func goFiles(dir string) []string {
	es, _ := os.ReadDir(dir)
	var out []string
	for _, e := range es {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

func subdirs(dir string) []string {
	es, _ := os.ReadDir(dir)
	var out []string
	for _, e := range es {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

// misplaced is one file --check objects to.
type misplaced struct {
	path   string // relative to the repository root
	reason string
	regen  bool // fixed by re-running the tool with no flags, not by --place
}

// isSplit reports whether the tree has been split (core_cards.txt exists).
func isSplit(effDir string) bool {
	_, err := os.Stat(filepath.Join(effDir, coreCardsFile))
	return err == nil
}

// check lists every file of the effects tree that is not where the layout
// puts it. Before the split it only validates the range table.
func check(root string) ([]misplaced, error) {
	effDir := filepath.Join(root, "server", effectsRel)
	tbl, err := readRangeTable(filepath.Join(effDir, shardsFile))
	if err != nil {
		return nil, err
	}
	if !isSplit(effDir) {
		return nil, nil
	}
	coreCards, err := readLines(filepath.Join(effDir, coreCardsFile))
	if err != nil {
		return nil, err
	}
	rel := func(parts ...string) string {
		return filepath.ToSlash(filepath.Join(append([]string{"server", effectsRel}, parts...)...))
	}
	var out []misplaced
	for _, n := range goFiles(effDir) {
		if strings.HasSuffix(n, "_test.go") {
			out = append(out, misplaced{path: rel(n), reason: "a test file in the core package; tests live in " + rel("tests", "<id>") + "/"})
			continue
		}
		if coreCards[n] {
			continue
		}
		init, err := hasInit(filepath.Join(effDir, n))
		if err != nil {
			return nil, err
		}
		if init {
			out = append(out, misplaced{path: rel(n), reason: "a card file (it has an init()) in the core package; cards live in " + rel("cards", "<shard>") + "/"})
		}
	}
	cardIDs := map[string]bool{}
	for _, r := range tbl.cards {
		cardIDs[r.id] = true
	}
	testIDs := map[string]bool{}
	for _, r := range tbl.tests {
		testIDs[r.id] = true
	}
	for _, id := range subdirs(filepath.Join(effDir, "cards")) {
		if !cardIDs[id] {
			out = append(out, misplaced{path: rel("cards", id) + "/", reason: "a shard directory that " + rel(shardsFile) + " does not list"})
			continue
		}
		for _, n := range goFiles(filepath.Join(effDir, "cards", id)) {
			p := filepath.Join(effDir, "cards", id, n)
			switch {
			case strings.HasSuffix(n, "_test.go"):
				out = append(out, misplaced{path: rel("cards", id, n), reason: "a test file in a card shard; tests live in " + rel("tests", "<id>") + "/"})
			case tbl.cardShard(n) != shardKey(id):
				out = append(out, misplaced{path: rel("cards", id, n), reason: "in shard " + id + " but " + shardsFile + " puts it in " + pkgDir(tbl.cardShard(n))})
			default:
				init, err := hasInit(p)
				if err != nil {
					return nil, err
				}
				if !init {
					out = append(out, misplaced{path: rel("cards", id, n), reason: "a file with no init() in a card shard; shared helpers live in the core package " + rel() + "/"})
				}
			}
		}
	}
	for _, id := range subdirs(filepath.Join(effDir, "tests")) {
		if !testIDs[id] {
			out = append(out, misplaced{path: rel("tests", id) + "/", reason: "a test package that " + rel(shardsFile) + " does not list"})
			continue
		}
		for _, n := range goFiles(filepath.Join(effDir, "tests", id)) {
			switch {
			case n == generatedMainTest:
			case !strings.HasSuffix(n, "_test.go"):
				out = append(out, misplaced{path: rel("tests", id, n), reason: "a non-test file in a test package; test helpers live in " + rel("effectstest") + "/"})
			case tbl.testPkg(n) != testKey(id):
				out = append(out, misplaced{path: rel("tests", id, n), reason: "in test package " + id + " but " + shardsFile + " puts it in " + pkgDir(tbl.testPkg(n))})
			}
		}
	}
	// The aggregator links exactly the table's shards.
	agg := filepath.Join(effDir, "cards", "cards.go")
	if f, err := parser.ParseFile(token.NewFileSet(), agg, nil, parser.ImportsOnly); err != nil {
		out = append(out, misplaced{path: rel("cards", "cards.go"), reason: "the aggregator is missing or does not parse", regen: true})
	} else {
		have := map[string]bool{}
		for _, im := range f.Imports {
			p, _ := strconv.Unquote(im.Path.Value)
			have[p] = true
		}
		for _, r := range tbl.cards {
			if !have[importPath(shardKey(r.id))] {
				out = append(out, misplaced{path: rel("cards", "cards.go"), reason: "the aggregator does not link shard " + r.id, regen: true})
			}
		}
	}
	for _, n := range goFiles(filepath.Join(effDir, "effectstest")) {
		if strings.HasSuffix(n, "_test.go") {
			out = append(out, misplaced{path: rel("effectstest", n), reason: "a test file in effectstest, which holds helpers only; tests live in " + rel("tests", "<id>") + "/"})
		}
	}
	return out, nil
}

// fixItMessage is what --check prints, and what the docsguard test shows.
func fixItMessage(ms []misplaced) string {
	var b strings.Builder
	b.WriteString("server/internal/cards/effects is split into a core, card shards and test packages (#5906, ADR 0225),\n")
	b.WriteString("and these files are not where that layout puts them:\n\n")
	var files []string
	regen := false
	for _, m := range ms {
		fmt.Fprintf(&b, "  %s\n      %s\n", m.path, m.reason)
		switch {
		case m.regen:
			regen = true
		case !strings.HasSuffix(m.path, "/"):
			files = append(files, m.path)
		}
	}
	if regen {
		b.WriteString("\nRegenerate the aggregator, from the repository root: scripts/split-effects.sh\n")
	}
	if len(files) > 0 {
		b.WriteString("\nFix it with this one command, run from the repository root (the directory that holds server/):\n\n")
		fmt.Fprintf(&b, "  %s %s\n\n", placeCmd, strings.Join(files, " "))
		b.WriteString("It moves each card file into its shard (internal/cards/effects/cards/<shard>/) and each test file into\n")
		b.WriteString("its test package (internal/cards/effects/tests/<id>/), and rewrites the helper names they use to the\n")
		b.WriteString("split's names (onBattlefield becomes effects.OnBattlefield, newCatalogGame becomes\n")
		b.WriteString("effectstest.NewCatalogGame). It also repairs any core file a rebase left with old names, and exports\n")
		b.WriteString("a core helper a card needs. Then build and test, and commit what it changed.\n")
		b.WriteString("docs/adding-cards.md, \"Where a card goes\", has the rules.\n")
	} else if !regen {
		b.WriteString("\nEdit server/internal/cards/effects/shards.txt or remove the directory; see docs/adding-cards.md, \"Where a card goes\".\n")
	}
	return b.String()
}
