package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
)

// arena_test.go covers the resolution rules — what the flags, the
// environment and the defaults add up to — because that is where a
// mistake produces a run that works and measures the wrong thing. The
// playing itself is tested in internal/botarena.

func TestParseArenaFlagsDefaults(t *testing.T) {
	t.Setenv("CMDCTRL_OPENAI_ENDPOINT", "")
	t.Setenv("CMDCTRL_BOT_MODEL", "")
	t.Setenv("CMDCTRL_BOT_MAX_THINK", "")
	t.Setenv("CMDCTRL_SCRYFALL_DUMP", "")

	a, err := parseArenaFlags([]string{"--seats", "heuristic,heuristic"}, io.Discard)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(a.seats) != 2 || a.seats[0] != tiers.Heuristic {
		t.Errorf("seats = %v", a.seats)
	}
	if len(a.decks) != 2 || a.decks[0] != "" {
		t.Errorf("no --decks should deal the battle deck to every seat, got %v", a.decks)
	}
	if a.games != 10 || a.seed != 1 || a.turns != 60 || a.wall != 30*time.Minute {
		t.Errorf("defaults moved: games=%d seed=%d turns=%d wall=%s", a.games, a.seed, a.turns, a.wall)
	}
	if a.needsModel || a.needsIndex {
		t.Errorf("a heuristic-only battle-deck run needs neither a model nor a dump (model=%v index=%v)", a.needsModel, a.needsIndex)
	}
	if a.maxThink != 0 {
		t.Errorf("no endpoint and no env: max-think should stay at the tier default, got %s", a.maxThink)
	}
	if !a.printMD || a.printJSON {
		t.Errorf("Markdown is the default output (md=%v json=%v)", a.printMD, a.printJSON)
	}
}

func TestParseArenaFlagsSeatsRequired(t *testing.T) {
	if _, err := parseArenaFlags(nil, io.Discard); err == nil {
		t.Fatal("a run with no seats was accepted")
	}
}

func TestParseArenaFlagsRejectsUnknownTier(t *testing.T) {
	_, err := parseArenaFlags([]string{"--seats", "heuristic,genius"}, io.Discard)
	if err == nil {
		t.Fatal("an unknown tier was accepted")
	}
	if !strings.Contains(err.Error(), "genius") {
		t.Errorf("the error should name the tier, got %v", err)
	}
}

// One deck per seat or none at all. A partial list is the
// configuration that silently measures nothing: two seats on curated
// decks and two on vanilla bears is not a comparison.
func TestParseArenaFlagsDeckArity(t *testing.T) {
	if _, err := parseArenaFlags([]string{"--seats", "heuristic,heuristic", "--decks", "izzet-aggro"}, io.Discard); err == nil {
		t.Fatal("a deck list shorter than the seat list was accepted")
	}
	a, err := parseArenaFlags([]string{"--seats", "heuristic,heuristic", "--decks", "izzet-aggro,simic-ramp"}, io.Discard)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !a.needsIndex {
		t.Error("curated decks need a Scryfall index")
	}
}

// The local-endpoint think default, mirrored from cmd/server/main.go:
// a self-hosted model cannot answer inside the tier's 2s, so a run
// that did not say otherwise would measure timeouts.
func TestParseArenaFlagsLocalThinkDefault(t *testing.T) {
	t.Setenv("CMDCTRL_BOT_MAX_THINK", "")
	a, err := parseArenaFlags([]string{"--seats", "assisted,heuristic", "--endpoint", "http://box:11434/v1"}, io.Discard)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if a.maxThink != localDefaultMaxThink || !a.thinkDefaulted {
		t.Errorf("max-think = %s (defaulted %v), want %s", a.maxThink, a.thinkDefaulted, localDefaultMaxThink)
	}
	if !a.needsModel {
		t.Error("an assisted seat needs a model")
	}

	// An explicit flag always wins.
	a, err = parseArenaFlags([]string{"--seats", "assisted,heuristic", "--endpoint", "http://box:11434/v1", "--max-think", "7s"}, io.Discard)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if a.maxThink != 7*time.Second || a.thinkDefaulted {
		t.Errorf("--max-think 7s was overridden: got %s (defaulted %v)", a.maxThink, a.thinkDefaulted)
	}
}

func TestParseArenaFlagsEnvFallbacks(t *testing.T) {
	t.Setenv("CMDCTRL_OPENAI_ENDPOINT", "http://env:11434/v1")
	t.Setenv("CMDCTRL_BOT_MODEL", "qwen3:14b")
	t.Setenv("CMDCTRL_BOT_FRONTIER_MODEL", "qwen3:32b")
	t.Setenv("CMDCTRL_BOT_MAX_THINK", "12s")
	t.Setenv("CMDCTRL_SCRYFALL_DUMP", "/tmp/default-cards.json")

	a, err := parseArenaFlags([]string{"--seats", "assisted,heuristic"}, io.Discard)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if a.endpoint != "http://env:11434/v1" || a.modelID != "qwen3:14b" || a.frontier != "qwen3:32b" {
		t.Errorf("env fallbacks did not apply: %+v", a)
	}
	if a.maxThink != 12*time.Second {
		t.Errorf("CMDCTRL_BOT_MAX_THINK = %s, want 12s", a.maxThink)
	}
	if a.dump != "/tmp/default-cards.json" {
		t.Errorf("dump = %q", a.dump)
	}

	// Flags beat the environment.
	a, err = parseArenaFlags([]string{"--seats", "assisted,heuristic", "--endpoint", "http://flag:1/v1", "--model", "flag-model"}, io.Discard)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if a.endpoint != "http://flag:1/v1" || a.modelID != "flag-model" {
		t.Errorf("flags did not beat the environment: %+v", a)
	}
}

// Artifacts need somewhere to go. Silently discarding a decision log
// the operator asked for is worse than refusing the run.
func TestParseArenaFlagsArtifactsNeedOut(t *testing.T) {
	for _, flagName := range []string{"--decision-log", "--replays"} {
		if _, err := parseArenaFlags([]string{"--seats", "heuristic,heuristic", flagName}, io.Discard); err == nil {
			t.Errorf("%s with no --out was accepted", flagName)
		}
	}
	if _, err := parseArenaFlags([]string{"--seats", "heuristic,heuristic", "--decision-log", "--out", "/tmp/x"}, io.Discard); err != nil {
		t.Errorf("--decision-log with --out: %v", err)
	}
}

// A single-tier run on four decks must be able to tally per chair,
// or it collapses into one row that says nothing.
func TestParseArenaFlagsNames(t *testing.T) {
	a, err := parseArenaFlags([]string{
		"--seats", "heuristic,heuristic",
		"--decks", "izzet-aggro,simic-ramp",
		"--names", "izzet,simic",
	}, io.Discard)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	cfg := a.config(nil, nil, nil, "", nil)
	if cfg.Seats[0].Label() != "izzet" || cfg.Seats[1].Label() != "simic" {
		t.Errorf("labels = %q, %q", cfg.Seats[0].Label(), cfg.Seats[1].Label())
	}
	if _, err := parseArenaFlags([]string{"--seats", "heuristic,heuristic", "--names", "only-one"}, io.Discard); err == nil {
		t.Error("a name list shorter than the seat list was accepted")
	}
}

func TestParseArenaFlagsRejectsBadDecisionLogMode(t *testing.T) {
	if _, err := parseArenaFlags([]string{"--seats", "heuristic,heuristic", "--decision-log-mode", "everything"}, io.Discard); err == nil {
		t.Fatal("an unknown decision-log mode was accepted")
	}
}

func TestParseArenaFlagsRejectsZeroGames(t *testing.T) {
	if _, err := parseArenaFlags([]string{"--seats", "heuristic,heuristic", "--games", "0"}, io.Discard); err == nil {
		t.Fatal("--games 0 was accepted")
	}
}

func TestParseArenaFlagsRejectsMultipleStdoutFormats(t *testing.T) {
	if _, err := parseArenaFlags([]string{"--seats", "heuristic,heuristic", "--md", "--json"}, io.Discard); err == nil {
		t.Fatal("--md and --json together were accepted")
	}
}

func TestCreateArenaRunDirNeverReusesADirectory(t *testing.T) {
	root := t.TempDir()
	started := time.Date(2026, time.September, 17, 18, 19, 4, 0, time.UTC)
	first, err := createArenaRunDir(root, started)
	if err != nil {
		t.Fatal(err)
	}
	second, err := createArenaRunDir(root, started)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatalf("two runs reused %s", first)
	}
	for _, dir := range []string{first, second} {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			t.Errorf("run directory %s was not created: %v", dir, err)
		}
		if filepath.Dir(dir) != root {
			t.Errorf("run directory %s escaped root %s", dir, root)
		}
	}
}

// The config a parsed command line produces is the one the arena
// runs, so the mapping is checked rather than assumed — including the
// refusal a model tier with no client gets.
func TestArenaFlagsConfig(t *testing.T) {
	a, err := parseArenaFlags([]string{
		"--seats", "assisted,heuristic,heuristic,heuristic",
		"--decks", "izzet-aggro,simic-ramp,esper-control,mono-black-aristocrats",
		"--games", "4", "--seed", "77", "--rotate", "--turn-budget", "45",
		"--endpoint", "http://box:11434/v1", "--model", "qwen3:14b",
		"--replays", "--out", "/tmp/arena",
	}, io.Discard)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	cfg := a.config(nil, nil, nil, "/tmp/arena/run", nil)
	if len(cfg.Seats) != 4 || cfg.Seats[0].Tier != tiers.Assisted || cfg.Seats[0].Deck != "izzet-aggro" {
		t.Fatalf("seats = %+v", cfg.Seats)
	}
	if cfg.Games != 4 || cfg.Seed != 77 || !cfg.Rotate || cfg.TurnBudget != 45 {
		t.Errorf("config = %+v", cfg)
	}
	if cfg.Models.Routine != "qwen3:14b" {
		t.Errorf("model id did not reach the config: %+v", cfg.Models)
	}
	if cfg.ReplayDir != "/tmp/arena/run" {
		t.Errorf("--replays should point the room at the run directory, got %q", cfg.ReplayDir)
	}
	// nil client, assisted seat: refused, not downgraded.
	err = cfg.Validate()
	if err == nil {
		t.Fatal("an assisted seat with no client was accepted")
	}
	if !strings.Contains(err.Error(), "endpoint") {
		t.Errorf("the refusal should say what to do, got %v", err)
	}
}

// A run that stopped early measured less than it was asked to, and an
// operator scripting these must not read exit 0 off it. A Ctrl-C is
// the operator's own choice and is not a failure.
func TestArenaExit(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"a complete run", nil, 0},
		{"a Ctrl-C between games", context.Canceled, 0},
		{"a wrapped Ctrl-C", fmt.Errorf("botarena: %w", context.Canceled), 0},
		{"a game that could not be started", errors.New("botarena: add player: deck is empty"), 1},
		{"a deadline the caller set", context.DeadlineExceeded, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := arenaExit(c.err); got != c.want {
				t.Errorf("arenaExit(%v) = %d, want %d", c.err, got, c.want)
			}
		})
	}
}

// The run directory and games.jsonl carry the same class of
// aggregated hidden information the decision log does: a stalled
// game's dump enumerates every seat's legal moves, cast moves
// included, so it names castable cards in all four hands. ADR 0052's
// modes, for ADR 0052's reason.
func TestArenaArtifactsArePrivate(t *testing.T) {
	root := t.TempDir()
	dir, err := createArenaRunDir(root, time.Date(2026, time.September, 17, 18, 19, 4, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o700 {
		t.Errorf("run directory %s is %#o, want 0700", dir, perm)
	}
	f, err := createGamesJSONL(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	fi, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Errorf("games.jsonl is %#o, want 0600", perm)
	}
	if got := filepath.Base(f.Name()); got != "games.jsonl" {
		t.Errorf("wrote %q, want games.jsonl", got)
	}
}

// Arena seats get production's block grace, because with MinThink at
// 0 an attacker without it re-steps on its own commit and races the
// defenders — every combat resolving with fewer blocks than the same
// policies would declare at a real table. An operator who would
// rather have the wall clock back can turn it off, and "off" has to
// survive the arena's zero-means-production default.
func TestParseArenaFlagsBlockGrace(t *testing.T) {
	a, err := parseArenaFlags([]string{"--seats", "heuristic,heuristic"}, io.Discard)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	want := aiseat.DefaultConfig().BlockGrace
	if a.blockGrace != want {
		t.Errorf("--block-grace defaults to %s, want production's %s", a.blockGrace, want)
	}
	if got := a.config(nil, nil, nil, "", nil).Runner.BlockGrace; got != want {
		t.Errorf("the config carries %s, want %s", got, want)
	}

	off, err := parseArenaFlags([]string{"--seats", "heuristic,heuristic", "--block-grace", "0"}, io.Discard)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := off.config(nil, nil, nil, "", nil).Runner.BlockGrace; got >= 0 {
		t.Errorf("--block-grace 0 must reach the arena as an explicit off (a negative duration), got %s", got)
	}

	custom, err := parseArenaFlags([]string{"--seats", "heuristic,heuristic", "--block-grace", "1500ms"}, io.Discard)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := custom.config(nil, nil, nil, "", nil).Runner.BlockGrace; got != 1500*time.Millisecond {
		t.Errorf("--block-grace 1500ms reached the arena as %s", got)
	}
}

// #1503: --lockstep reaches the arena, and leaving it off leaves the
// arena on its concurrent default — one runner goroutine per seat, the
// schedule a live table runs.
func TestParseArenaFlagsLockstep(t *testing.T) {
	def, err := parseArenaFlags([]string{"--seats", "heuristic,heuristic"}, io.Discard)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if def.lockstep || def.config(nil, nil, nil, "", nil).Lockstep {
		t.Error("a run without --lockstep must stay on the concurrent schedule")
	}
	on, err := parseArenaFlags([]string{"--seats", "heuristic,heuristic", "--lockstep"}, io.Discard)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !on.config(nil, nil, nil, "", nil).Lockstep {
		t.Error("--lockstep did not reach the arena's config")
	}
}
