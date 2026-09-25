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
// — and lists every ROUTE to a func or an interface (which can hold
// one): every field of every struct reachable from Game whose type
// reaches one, directly or through another struct (#1558). It compares
// that list with testdata/closure_fields.txt. The drift test classifies
// seven types by field name; this one follows what those fields HOLD,
// all the way down, which is where a closure hides.
//
// THE RATCHET, in two halves.
//
//   - A new route fails the build until it is classified — including a
//     new field whose type is a struct that is already listed, which is
//     the hole #1558 closed (a `[]ReplacementEffect` on TargetSpec used
//     to arrive under ReplacementEffect's existing lines). A `census:`
//     line may only name a live counter, and never one listed in
//     retiredCensusCounters — so a tier that retires a counter makes
//     every route still charged to it fail.
//   - The number of lines in each blocker class (every `census:<C>` and
//     `transient`) may only FALL: closureClassCeilings pins it, and a
//     count above its ceiling fails. So does a count below it, until
//     the ceiling is lowered to match — that is the ratchet clicking.
//     Tier PRs delete lines and lower ceilings; nothing raises one
//     without a reviewer reading why in the diff.
//   - ADR 0041 P11: a retirement may move a line to the class of its
//     remaining route, in the same PR, and the PR lists the lines and
//     the new ceiling. A field reachable by several routes carries the
//     most restrictive class among them, so deleting the route that set
//     it lets it fall to the next one — usually the resume frames. That
//     is not a new route, and it is the one way a ceiling goes up.
//
// Owner decision 4 (2026-09-24): ChoiceResumeFrames is the one census
// counter that stays allowed after tier 4 — resume frames are out of
// scope for this sprint. Its lines are charged to it deliberately.
//
// To regenerate after a deliberate change:
//
//	go test ./internal/game -run TestClosureFieldsReachableFromGame -args -update-closure-fields
//
// which keeps every existing classification and marks new routes
// `unclassified`, which still fails until a human classifies them. It
// never touches closureClassCeilings.

var updateClosureFields = flag.Bool("update-closure-fields", false,
	"rewrite testdata/closure_fields.txt, keeping classifications and marking new paths unclassified")

const closureFieldsFile = "closure_fields.txt"

// retiredCensusCounters are the counters a landed phase-3 tier has
// retired. A closure field charged to one of them fails the build.
var retiredCensusCounters = map[string]string{
	"DelayedTriggerEffects":  "ADR 0041 phase 3 tier 2 (#1497): a delayed trigger is a registered body key plus plain params",
	"ScopedStatics":          "ADR 0041 phase 3 tier 3a (#1497): a continuous effect with a duration is a ScopedEffect record over the closed Mod vocabulary",
	"StackTargetSpecs":       "ADR 0041 phase 3 tier 4 (#1497, P9): a stack item's target and mode clauses are re-derived from its oracle ID or its catalog ability ref, and an item that can be neither is counted once, in StackEffects",
	"TurnScopedReplacements": "ADR 0041 phase 3 tier 3b (#1497): a replacement effect a spell creates is a ScopedEffect record with a replacement-reader mod",
}

// closureClassCeilings is the ratchet's second half (#1558): how many
// lines of testdata/closure_fields.txt each blocker class may hold. A
// class missing from the map has a ceiling of zero. The counts may
// only fall — lower the number in the same PR that deletes the lines,
// and never raise one to make a new route pass: classify the route as
// what it really is, or make it data.
//
// One exception, which is not a new route (ADR 0041 P11): a retirement
// may move a line to the class of its REMAINING route, in the same PR,
// and the PR lists the lines and the new ceiling. A field reachable by
// several routes carries the most restrictive class among them, so
// deleting the route that set its class drops it to the next one —
// usually a paused prompt's resume frame. Tier 4-1 moved eleven lines
// that way and tier 3b-1 nine (ReplacementEffect.*, CopySelector.*,
// EntryHandReveal.*): ChoiceResumeFrames 98 → 118.
var closureClassCeilings = map[string]int{
	// 98 + the 11 census:StackTargetSpecs lines tier 4's first slice
	// moved here under ADR 0041 P11 (ModeOption.Effect,
	// ModeOption.Targets, ModeSpec.Options, StackItem.modeSpec,
	// StackItem.targetSpec, TargetDifference.Key and
	// TargetSpec.{AbilityOK, CardOK, Different, PlayerOK, Rest}) + the 9
	// census:TurnScopedReplacements lines tier 3b-1 moved
	// (ReplacementEffect.*, CopySelector.*, EntryHandReveal.*).
	"census:ChoiceResumeFrames":    118,
	"census:IntrinsicAbilityCards": 45,
	"census:StackEffects":          6,
	"census:TurnScopedBlockRules":  4,
	"transient":                    1,
}

// isCeilingedClass reports whether a class is one closureClassCeilings
// holds to a count: every blocker, and transient (which is a claim a
// reviewer cannot check at a glance, so it is not free either).
func isCeilingedClass(class string) bool {
	return class == "transient" || strings.HasPrefix(class, "census:")
}

// checkClosureClassCeilings compares the per-class line counts of a
// classification with closureClassCeilings.
func checkClosureClassCeilings(recorded map[string]string, ceilings map[string]int) []string {
	counts := map[string]int{}
	for _, class := range recorded {
		if isCeilingedClass(class) {
			counts[class]++
		}
	}
	var problems []string
	for class, n := range counts {
		ceiling := ceilings[class]
		if n > ceiling {
			problems = append(problems, fmt.Sprintf(
				"testdata/%s has %d %s lines; closureClassCeilings allows %d. A blocker class may only shrink: classify the new route as what it really is (rebuilt, keyed), or make the closure data — do not raise the ceiling",
				closureFieldsFile, n, class, ceiling))
		}
		if n < ceiling {
			problems = append(problems, fmt.Sprintf(
				"testdata/%s has %d %s lines, below its ceiling of %d — lower closureClassCeilings[%q] to %d (that is the ratchet clicking)",
				closureFieldsFile, n, class, ceiling, class, n))
		}
	}
	for class, ceiling := range ceilings {
		if counts[class] == 0 && ceiling != 0 {
			problems = append(problems, fmt.Sprintf(
				"testdata/%s has no %s lines left — delete its closureClassCeilings entry",
				closureFieldsFile, class))
		}
	}
	sort.Strings(problems)
	return problems
}

// closureField is one line of the ratchet: a field that can hold a
// closure, and the first route by which Game reaches it (for the
// reviewer; not part of the key).
type closureField struct {
	path string // "<Type>.<Field>", or "Game.<Field>" for a registry
	via  string // "Game.StackMeta{}.Effect" — first route found
}

// closureFieldPaths walks the type graph from Game breadth-first and
// returns one line per ROUTE: "<Type>.<Field>" for every field of every
// game struct reachable from Game (Game itself included, as
// "Game.<Field>") whose type can reach a func or an interface — held
// directly (through pointers, slices, arrays and maps) or through
// another struct that holds one, all the way down.
//
// Keying by route rather than by the closure-holding TYPE is the point
// (#1558): a new field whose type is an already-listed struct — a
// `[]ReplacementEffect` on TargetSpec, a `[]StaticAbility` on Card —
// is a new way for a closure to reach a restore point, and it now
// fails until somebody classifies it. Keyed by type, it arrived under
// the existing ReplacementEffect / StaticAbility lines and passed
// every guard.
func closureFieldPaths() []closureField {
	return closureRoutesFrom(reflect.TypeOf(Game{}))
}

// closureRoutesFrom is closureFieldPaths over any root struct of this
// package, so the walker itself can be tested on types built to probe
// it (closure_fields_probe_test.go).
func closureRoutesFrom(root reflect.Type) []closureField {
	pkg := root.PkgPath()

	// direct reports whether t holds a func/interface without passing
	// through a named game struct, and returns the named game structs
	// it passes to. An anonymous struct is part of the field that
	// declares it, so its fields are walked in place.
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
			if t.Name() == "" {
				hit := false
				var out []reflect.Type
				for i := 0; i < t.NumField(); i++ {
					d, ss := direct(t.Field(i).Type)
					hit = hit || d
					out = append(out, ss...)
				}
				return hit, out
			}
			if t.PkgPath() == pkg {
				return false, []reflect.Type{t}
			}
		}
		return false, nil
	}

	// Every game struct reachable from Game, then reachability as a
	// fixed point over that graph. (A memoised DFS that treats a type
	// on the current path as false memoises the wrong answer for the
	// types inside a cycle, and a missed type is a missed route.)
	gameT := root
	children := map[reflect.Type][]reflect.Type{}
	reaches := map[reflect.Type]bool{}
	var order []reflect.Type
	{
		seen := map[reflect.Type]bool{gameT: true}
		queue := []reflect.Type{gameT}
		for len(queue) > 0 {
			t := queue[0]
			queue = queue[1:]
			order = append(order, t)
			for i := 0; i < t.NumField(); i++ {
				d, ss := direct(t.Field(i).Type)
				if d {
					reaches[t] = true
				}
				for _, s := range ss {
					children[t] = append(children[t], s)
					if !seen[s] {
						seen[s] = true
						queue = append(queue, s)
					}
				}
			}
		}
	}
	for changed := true; changed; {
		changed = false
		for _, t := range order {
			if reaches[t] {
				continue
			}
			for _, c := range children[t] {
				if reaches[c] {
					reaches[t] = true
					changed = true
					break
				}
			}
		}
	}
	fieldReaches := func(ft reflect.Type) bool {
		d, ss := direct(ft)
		if d {
			return true
		}
		for _, s := range ss {
			if reaches[s] {
				return true
			}
		}
		return false
	}

	// The routes, breadth-first so each line's `via` is the shortest
	// path Game reaches it by.
	found := map[string]closureField{}
	type item struct {
		t   reflect.Type
		via string
	}
	seen := map[reflect.Type]bool{gameT: true}
	queue := []item{{gameT, gameT.Name()}}
	for len(queue) > 0 {
		it := queue[0]
		queue = queue[1:]
		for i := 0; i < it.t.NumField(); i++ {
			f := it.t.Field(i)
			if !fieldReaches(f.Type) {
				continue
			}
			via := it.via + "." + f.Name
			p := it.t.Name() + "." + f.Name
			if _, dup := found[p]; !dup {
				found[p] = closureField{path: p, via: via}
			}
			_, structs := direct(f.Type)
			for _, s := range structs {
				if !seen[s] && reaches[s] {
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
			t.Errorf(`%s reaches a func or an interface and is not in testdata/%s.

A new route from Game reaches something that can hold a closure — a new
field, or a new field whose type is a struct that is already listed. ADR 0041 phase 3
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
			t.Errorf("testdata/%s lists %s, which no longer reaches a func or an interface — delete the line (that is the ratchet working)", closureFieldsFile, p)
		}
	}
	for _, problem := range checkClosureClassCeilings(recorded, closureClassCeilings) {
		t.Error(problem)
	}
}

const closureFieldsHeader = `# closure_fields.txt — ADR 0041 phase 3's ratchet (Decision P3, #1497).
#
# Every ROUTE from Game to a func or an interface — every field of
# every struct reachable from Game whose type reaches one, directly or
# through another struct (#1558) — and why it is allowed to:
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
# A line is one edge, <Type>.<Field>. Where Game reaches the SAME type
# by several routes the edges into it differ (Game.BuiltinReplacements
# is rebuilt, a paused prompt's replacementResume is a blocker), and the lines
# for that type's own fields carry the most restrictive class of every
# route in: ReplacementEffect.Replace is charged to the blocker until
# the tier that retires it. The trailing comment is the shortest route.
#
# Tier PRs DELETE lines, and lower closureClassCeilings in
# closure_fields_test.go to match: every census: and transient class is
# held to a count that may only fall. A counter a tier retires
# (retiredCensusCounters) fails every line still charged to it.
#
# Owner decision 4 (2026-09-24): census:ChoiceResumeFrames is the one
# counter that stays allowed after tier 4 — resume frames are out of
# scope for this sprint.
#
# Regenerate (keeps classifications, marks new routes unclassified):
#   go test ./internal/game -run TestClosureFieldsReachableFromGame -args -update-closure-fields

`
