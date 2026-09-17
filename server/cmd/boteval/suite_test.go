package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// suite_test.go pins the three subcommands' flag surface. It is the
// part of a CLI that silently rots: a renamed flag or a changed
// default turns a documented command line into a run that measured
// something else, and nothing else in the tree reads these.

func TestParseSuiteRunDefaults(t *testing.T) {
	o, err := parseSuiteRun(nil)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if o.Dir != defaultPositionsDir {
		t.Errorf("--dir default %q, want %q", o.Dir, defaultPositionsDir)
	}
	if o.Policy != "heuristic" {
		t.Errorf("--policy default %q, want heuristic", o.Policy)
	}
	if o.Parallel != 1 {
		t.Errorf("--parallel default %d, want 1 (a single-GPU endpoint is slower with more in flight, not faster)", o.Parallel)
	}
	if o.MD {
		t.Error("--md defaults on")
	}
}

func TestParseSuiteRun(t *testing.T) {
	o, err := parseSuiteRun([]string{
		"--dir", "pos", "--policy", "assisted", "--deck", "dimir-mill",
		"--max-think", "20s", "--parallel", "4", "--out", "r.json", "--md",
	})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if o.Dir != "pos" || o.Policy != "assisted" || o.Deck != "dimir-mill" ||
		o.MaxThink != 20*time.Second || o.Parallel != 4 || o.Out != "r.json" || !o.MD {
		t.Fatalf("parsed %+v", o)
	}
}

func TestParseSuiteRunRejectsNonsense(t *testing.T) {
	for _, args := range [][]string{
		{"--policy", "genius"}, // not a tier
		{"--parallel", "0"},    // would deadlock the worker pool
	} {
		if _, err := parseSuiteRun(args); err == nil {
			t.Errorf("%v was accepted", args)
		}
	}
}

// The deployment's own deadline is the fallback for --max-think: a
// local model is routinely slower than the tier default, and the suite
// has to ask the question the live seat would.
func TestSuiteRunTakesMaxThinkFromTheEnvironment(t *testing.T) {
	t.Setenv("CMDCTRL_BOT_MAX_THINK", "45s")
	o, err := parseSuiteRun(nil)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if o.MaxThink != 45*time.Second {
		t.Fatalf("--max-think %v, want 45s from CMDCTRL_BOT_MAX_THINK", o.MaxThink)
	}
}

func TestParseSuiteHarvest(t *testing.T) {
	o, err := parseSuiteHarvest([]string{
		"--from", "a.jsonl,b.jsonl", "--to", "inbox",
		"--escalated", "--disagree",
		"--fallback", "malformed,timeout", "--layer", "B,C",
		"--seat", "0,2", "--tag", "block,attack",
		"--limit", "12", "--seed", "9",
		"extra.jsonl",
	})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got, want := o.From, []string{"a.jsonl", "b.jsonl", "extra.jsonl"}; !eq(got, want) {
		t.Errorf("--from %v, want %v", got, want)
	}
	if o.To != "inbox" || !o.Filter.Escalated || !o.Filter.Disagree {
		t.Errorf("parsed %+v", o)
	}
	if !eq(o.Filter.Fallbacks, []string{"malformed", "timeout"}) || !eq(o.Filter.Layers, []string{"B", "C"}) ||
		!eq(o.Filter.Tags, []string{"block", "attack"}) {
		t.Errorf("filters %+v", o.Filter)
	}
	if len(o.Filter.Seats) != 2 || o.Filter.Seats[0] != 0 || o.Filter.Seats[1] != 2 {
		t.Errorf("--seat %v", o.Filter.Seats)
	}
	if o.Filter.Limit != 12 || o.Filter.Seed != 9 {
		t.Errorf("--limit/--seed %d/%d", o.Filter.Limit, o.Filter.Seed)
	}
}

func TestParseSuiteHarvestNeedsSourceAndDestination(t *testing.T) {
	if _, err := parseSuiteHarvest([]string{"--to", "inbox"}); err == nil {
		t.Error("a harvest with no --from was accepted")
	}
	if _, err := parseSuiteHarvest([]string{"--from", "a.jsonl"}); err == nil {
		t.Error("a harvest with no --to was accepted")
	}
	if _, err := parseSuiteHarvest([]string{"--from", "a.jsonl", "--to", "x", "--seat", "left"}); err == nil {
		t.Error("--seat left was accepted")
	}
}

func TestParseSuiteRender(t *testing.T) {
	o, err := parseSuiteRender([]string{"--pos", "p.json", "--deck", "izzet-aggro"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if o.Pos != "p.json" || o.Deck != "izzet-aggro" {
		t.Fatalf("parsed %+v", o)
	}
	// A bare path is the shape anybody types after a harvest prints
	// one, so it is accepted too.
	if o, err = parseSuiteRender([]string{"p.json"}); err != nil || o.Pos != "p.json" {
		t.Fatalf("bare path: %+v %v", o, err)
	}
	if _, err := parseSuiteRender(nil); err == nil {
		t.Error("a render with no position was accepted")
	}
}

// expandLogs is what turns the documented `--from 'dir/*.jsonl'` into
// files, and it also accepts a directory, because that is what an
// operator has after CMDCTRL_BOT_DECISION_LOG.
func TestExpandLogs(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.decisions.jsonl", "b.decisions.jsonl", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("{}\n"), 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	got, err := expandLogs([]string{dir})
	if err != nil {
		t.Fatalf("expand: %v", err)
	}
	want := []string{filepath.Join(dir, "a.decisions.jsonl"), filepath.Join(dir, "b.decisions.jsonl")}
	if !eq(got, want) {
		t.Fatalf("expanded to %v, want %v", got, want)
	}
	// A directory and an explicit file naming the same log must not
	// harvest it twice.
	got, err = expandLogs([]string{dir, want[0], filepath.Join(dir, "*.decisions.jsonl")})
	if err != nil {
		t.Fatalf("expand: %v", err)
	}
	if !eq(got, want) {
		t.Fatalf("deduped to %v, want %v", got, want)
	}
}

// A model tier with no transport is REFUSED rather than quietly run as
// Layer A + B: a report labelled `assisted` that measured the
// heuristic is a false measurement (tiers.Factory.TierStatus's rule).
func TestSuiteRunRefusesAModelTierWithNoClient(t *testing.T) {
	t.Setenv("CMDCTRL_OPENAI_ENDPOINT", "")
	t.Setenv("CMDCTRL_ANTHROPIC_API_KEY", "")
	t.Setenv("CMDCTRL_BOT_MODEL", "")
	_, err := buildSuitePolicy(os.Stdout, suiteRunOpts{Policy: "assisted"})
	if err == nil {
		t.Fatal("the assisted tier was built with no model transport")
	}
	// And the free tiers still build.
	if _, err := buildSuitePolicy(os.Stdout, suiteRunOpts{Policy: "heuristic"}); err != nil {
		t.Fatalf("heuristic: %v", err)
	}
}

func eq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
