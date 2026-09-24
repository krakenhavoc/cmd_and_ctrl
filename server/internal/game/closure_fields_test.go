package game

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// closure_fields_test.go is ADR 0041 phase 3's ratchet (Decision P3,
// #1497). After phase 3 a function value reachable from *Game is
// acceptable only if it is one of three kinds:
//
//	rebuilt    installed by NewGame, or derived by the running binary
//	           from carried data (listeners, built-in replacements, the
//	           ScopedEffect layer adapter)
//	keyed      looked up in a registry by a carried key (a catalog
//	           ability by CatalogKey, a token template, a grant bundle)
//	transient  cannot be alive at a capture point — created and used up
//	           inside one engine call, so it never outlives a Room.Apply
//	test-only  a test injection slot production has no path to
//
// Anything else is a blocker, and a blocker is named by the
// ContinuationCensus counter that accounts for it: `census:<Counter>`.
//
// The test walks the type graph from Game by reflection — struct
// fields, slices, arrays, maps, pointers, named types, exported or not
// — and lists every field whose type holds a func or an interface
// (which can hold one). It compares that list with
// testdata/closure_fields.txt. The drift test classifies seven types
// by field name; this one follows what those fields HOLD, all the way
// down, which is where a closure hides.
//
// THE RATCHET. A new path fails the build until it is classified. A
// `census:` line may only name a live counter, and never one listed in
// retiredCensusCounters — so a tier that retires a counter makes every
// path still charged to it fail, and nothing may add a `census:` line
// for it again. Tier PRs delete lines; they do not add `census:` ones.
//
// Owner decision 4 (2026-09-24): ChoiceResumeFrames is the one census
// counter that stays allowed after tier 4 — resume frames are out of
// scope for this sprint. Its lines are charged to it deliberately.
//
// To regenerate after a deliberate change:
//
//	go test ./internal/game -run TestClosureFieldsReachableFromGame -args -update-closure-fields
//
// which keeps every existing classification and marks new paths
// `unclassified`, which still fails until a human classifies them.

var updateClosureFields = flag.Bool("update-closure-fields", false,
	"rewrite testdata/closure_fields.txt, keeping classifications and marking new paths unclassified")

const closureFieldsFile = "closure_fields.txt"

// retiredCensusCounters are the counters a landed phase-3 tier has
// retired. A closure field charged to one of them fails the build.
// Empty until tier 2 retires DelayedTriggerEffects.
var retiredCensusCounters = map[string]string{}

// closureField is one line of the ratchet: a field that can hold a
// closure, and the first route by which Game reaches it (for the
// reviewer; not part of the key).
type closureField struct {
	path string // "<Type>.<Field>", or "Game.<Field>" for a registry
	via  string // "Game.StackMeta{}.Effect" — first route found
}

// closureFieldPaths walks the type graph from Game breadth-first and
// returns two kinds of line:
//
//   - "Game.<Field>" for every field of Game that REACHES a func or an
//     interface anywhere below it — the registry-level view, which is
//     where a tier's work shows (Game.ScopedEffects is absent because
//     it holds none; Game.ScopedStatics is present);
//   - "<Type>.<Field>" for every field of every game struct reachable
//     from Game whose own type holds a func or an interface directly
//     (through pointers, slices, arrays and maps, but not through
//     another struct) — the field-level view, which is where a new
//     closure appears.
func closureFieldPaths() []closureField {
	pkg := reflect.TypeOf(Game{}).PkgPath()

	// direct reports whether t holds a func/interface without passing
	// through a struct, and returns the game structs it passes to.
	var direct func(t reflect.Type) (bool, []reflect.Type)
	direct = func(t reflect.Type) (bool, []reflect.Type) {
		switch t.Kind() {
		case reflect.Func, reflect.Interface:
			return true, nil
		case reflect.Ptr, reflect.Slice, reflect.Array, reflect.Chan:
			return direct(t.Elem())
		case reflect.Map:
			k, ks := direct(t.Key())
			v, vs := direct(t.Elem())
			return k || v, append(ks, vs...)
		case reflect.Struct:
			if t.PkgPath() == pkg {
				return false, []reflect.Type{t}
			}
		}
		return false, nil
	}

	// reaches is the transitive version, memoised; a type on the
	// current path counts as false (a cycle adds nothing new).
	memo := map[reflect.Type]bool{}
	inProgress := map[reflect.Type]bool{}
	var reaches func(t reflect.Type) bool
	reaches = func(t reflect.Type) bool {
		if v, ok := memo[t]; ok {
			return v
		}
		if inProgress[t] {
			return false
		}
		inProgress[t] = true
		defer delete(inProgress, t)
		out := false
		for i := 0; i < t.NumField(); i++ {
			d, structs := direct(t.Field(i).Type)
			if d {
				out = true
			}
			for _, s := range structs {
				if reaches(s) {
					out = true
				}
			}
		}
		memo[t] = out
		return out
	}

	found := map[string]closureField{}
	gameT := reflect.TypeOf(Game{})
	for i := 0; i < gameT.NumField(); i++ {
		f := gameT.Field(i)
		d, structs := direct(f.Type)
		hit := d
		for _, s := range structs {
			if reaches(s) {
				hit = true
			}
		}
		if hit {
			p := "Game." + f.Name
			found[p] = closureField{path: p, via: p}
		}
	}

	type item struct {
		t   reflect.Type
		via string
	}
	seen := map[reflect.Type]bool{gameT: true}
	queue := []item{{gameT, "Game"}}
	for len(queue) > 0 {
		it := queue[0]
		queue = queue[1:]
		for i := 0; i < it.t.NumField(); i++ {
			f := it.t.Field(i)
			via := it.via + "." + f.Name
			d, structs := direct(f.Type)
			if d && it.t != gameT {
				p := it.t.Name() + "." + f.Name
				if _, dup := found[p]; !dup {
					found[p] = closureField{path: p, via: via}
				}
			}
			for _, s := range structs {
				if !seen[s] {
					seen[s] = true
					queue = append(queue, item{s, via})
				}
			}
		}
	}

	out := make([]closureField, 0, len(found))
	for _, f := range found {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].path < out[j].path })
	return out
}
func readClosureFields(t *testing.T) map[string]string {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", closureFieldsFile))
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}
		}
		t.Fatal(err)
	}
	defer f.Close()
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			t.Fatalf("%s: malformed line %q — want `<Type>.<Field> <class>`", closureFieldsFile, line)
		}
		out[fields[0]] = fields[1]
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestClosureFieldsReachableFromGame(t *testing.T) {
	live := closureFieldPaths()
	recorded := readClosureFields(t)

	if *updateClosureFields {
		var b strings.Builder
		b.WriteString(closureFieldsHeader)
		for _, f := range live {
			class := recorded[f.path]
			if class == "" {
				class = "unclassified"
			}
			fmt.Fprintf(&b, "%-44s %-34s # %s\n", f.path, class, f.via)
		}
		if err := os.WriteFile(filepath.Join("testdata", closureFieldsFile), []byte(b.String()), 0o644); err != nil {
			t.Fatal(err)
		}
		recorded = readClosureFields(t)
	}

	counters := map[string]bool{}
	ct := reflect.TypeOf(ContinuationCensus{})
	for i := 0; i < ct.NumField(); i++ {
		counters[ct.Field(i).Name] = true
	}

	liveSet := map[string]bool{}
	for _, f := range live {
		p := f.path
		liveSet[p] = true
		class, ok := recorded[p]
		if !ok {
			t.Errorf(`%s holds a func or an interface and is not in testdata/%s.

A field reachable from Game can now hold a closure. ADR 0041 phase 3
(Decision P3) allows only three kinds: rebuilt, keyed, transient. If it
is none of those it is a restore-point blocker, and it must be charged
to the ContinuationCensus counter that counts it (census:<Counter>).
Classify it, or run with -args -update-closure-fields and then do.`, p, closureFieldsFile)
			continue
		}
		switch {
		case class == "rebuilt", class == "keyed", class == "transient", class == "test-only":
		case strings.HasPrefix(class, "census:"):
			counter := strings.TrimPrefix(class, "census:")
			if !counters[counter] {
				t.Errorf("%s is charged to census:%s, which is not a ContinuationCensus counter", p, counter)
			}
			if why, retired := retiredCensusCounters[counter]; retired {
				t.Errorf("%s is charged to census:%s, which phase 3 has retired (%s). The closure has to become data, not a blocker.", p, counter, why)
			}
		default:
			t.Errorf("%s is classified %q; want rebuilt, keyed, transient, test-only or census:<Counter>", p, class)
		}
	}
	// The point of phase 3's records: they reach no closure at all.
	for _, dataOnly := range []string{"Game.ScopedEffects"} {
		if liveSet[dataOnly] {
			t.Errorf("%s reaches a func or an interface; ADR 0041 phase 3's records must be data all the way down", dataOnly)
		}
	}
	for p := range recorded {
		if !liveSet[p] {
			t.Errorf("testdata/%s lists %s, which no longer holds a func or an interface — delete the line (that is the ratchet working)", closureFieldsFile, p)
		}
	}
}

const closureFieldsHeader = `# closure_fields.txt — ADR 0041 phase 3's ratchet (Decision P3, #1497).
#
# Every field reachable from Game whose type holds a func or an
# interface, and why it is allowed to:
#
#   rebuilt            installed by NewGame, or rebuilt by the running
#                      binary from carried data
#   keyed              looked up in a registry by a carried key
#   transient          never alive at a capture point (the end of a
#                      Room.Apply)
#   census:<Counter>   a restore-point blocker, counted by that
#                      ContinuationCensus counter
#   test-only          a test injection slot production has no path to
#
# A TYPE-level line (<Type>.<Field>) carries the most restrictive
# class of every route Game reaches that type by: ReplacementEffect is
# rebuilt through BuiltinReplacements but a blocker through
# TurnScopedReplacements, so it is charged to the blocker until the
# tier that retires it. The trailing comment is the first route found.
#
# Tier PRs DELETE lines. Nothing adds a census: line; a counter a tier
# retires (retiredCensusCounters in closure_fields_test.go) fails every
# line still charged to it.
#
# Owner decision 4 (2026-09-24): census:ChoiceResumeFrames is the one
# counter that stays allowed after tier 4 — resume frames are out of
# scope for this sprint.
#
# Regenerate (keeps classifications, marks new paths unclassified):
#   go test ./internal/game -run TestClosureFieldsReachableFromGame -args -update-closure-fields

`
