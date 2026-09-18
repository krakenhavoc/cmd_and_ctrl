package botarena_test

import (
	"context"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/decisionlog"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // catalog hooks, so the deck's burn spell resolves
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// arena_test.go is in two halves.
//
// The ungated half tests the arithmetic — seat rotation, the Wilson
// interval, and that the report renders — because those are where a
// wrong answer is invisible. A rotation that misses a chair, or an
// interval computed with the wrong denominator, produces a report
// that looks exactly like a correct one and is wrong in the direction
// nobody checks.
//
// The gated half plays real games, and is gated for the usual reason:
// a two-seat heuristic game is a few seconds, which is a few seconds
// too many on every `go test ./...`.

func requireGameTests(t *testing.T) {
	t.Helper()
	if os.Getenv("AISEAT_GAME_TESTS") == "" {
		t.Skip("whole-game test: set AISEAT_GAME_TESTS=1 to run (the nightly does)")
	}
}

// --- rotation --------------------------------------------------------

// Every contestant must sit in every chair the same number of times
// over a run whose length is a multiple of the table size. That is
// the whole justification for rotation: turn order in Commander is
// worth real percentage points, so an unbalanced seating measures the
// chair rather than the policy.
func TestRotationCoversEveryPosition(t *testing.T) {
	for _, n := range []int{2, 3, 4} {
		games := 3 * n
		// counts[spec][position]
		counts := make([][]int, n)
		for i := range counts {
			counts[i] = make([]int, n)
		}
		for i := 0; i < games; i++ {
			order := botarena.Order(n, i, true)
			seen := map[int]bool{}
			for pos, spec := range order {
				if seen[spec] {
					t.Fatalf("n=%d game=%d: order %v seats spec %d twice", n, i, order, spec)
				}
				seen[spec] = true
				counts[spec][pos]++
			}
		}
		want := games / n
		for spec := range counts {
			for pos, got := range counts[spec] {
				if got != want {
					t.Errorf("n=%d: spec %d sat at position %d %d times, want %d (%v)", n, spec, pos, got, want, counts[spec])
				}
			}
		}
	}
}

// Rotation off must seat everybody where they were configured, every
// game — the mode for isolating a single change on identical deals.
func TestNoRotationKeepsSeating(t *testing.T) {
	for i := 0; i < 5; i++ {
		got := botarena.Order(4, i, false)
		for pos, spec := range got {
			if pos != spec {
				t.Fatalf("game %d: order %v is not the identity", i, got)
			}
		}
	}
}

// --- Wilson -----------------------------------------------------------

func TestWilson(t *testing.T) {
	const tol = 5e-4
	cases := []struct {
		wins, n        int
		wantLo, wantHi float64
	}{
		// The textbook example: half of ten is genuinely uncertain.
		{5, 10, 0.2366, 0.7634},
		// The case a normal approximation gets wrong, and the reason
		// Wilson is here: zero wins from ten games has NOT ruled out
		// a 25% policy, and the interval has to say so.
		{0, 10, 0.0000, 0.2775},
		{10, 10, 0.7225, 1.0000},
		{1, 4, 0.0455, 0.6997},
		{25, 100, 0.1755, 0.3432},
	}
	for _, c := range cases {
		lo, hi := botarena.Wilson(c.wins, c.n, botarena.DefaultZ)
		if math.Abs(lo-c.wantLo) > tol || math.Abs(hi-c.wantHi) > tol {
			t.Errorf("Wilson(%d, %d) = [%.4f, %.4f], want [%.4f, %.4f]", c.wins, c.n, lo, hi, c.wantLo, c.wantHi)
		}
		if lo < 0 || hi > 1 || lo > hi {
			t.Errorf("Wilson(%d, %d) = [%.4f, %.4f] is not a valid interval", c.wins, c.n, lo, hi)
		}
	}
	// No games is no evidence, and must read as the whole range
	// rather than as a confident zero.
	if lo, hi := botarena.Wilson(0, 0, botarena.DefaultZ); lo != 0 || hi != 1 {
		t.Errorf("Wilson(0, 0) = [%v, %v], want the whole range", lo, hi)
	}
}

// --- the report -------------------------------------------------------

func TestSummaryMarkdown(t *testing.T) {
	sum := botarena.Summary{
		Config: botarena.ConfigSummary{
			Seats: []botarena.SeatSpec{
				{Tier: tiers.Assisted, Deck: "izzet-aggro"},
				{Tier: tiers.Heuristic, Deck: "simic-ramp"},
			},
			Games: 2, Seed: 7, Rotate: true, TurnBudget: 60,
			MaxThink: 20 * time.Second, Routine: "qwen3:14b", Endpoint: "http://box:11434/v1",
			Note:   "q4_k_m",
			Chairs: botarena.ChairCounts(2, 2, true),
		},
		Started: time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC),
		Elapsed: 90 * time.Second,
		Games: []botarena.GameDigest{
			{Seed: 7, Turns: 21, State: game.StateEnded, Winner: "assisted", Elapsed: 45 * time.Second},
			{Seed: 8, Turns: 19, State: game.StateEnded, Winner: "heuristic", Elapsed: 45 * time.Second},
		},
		PerPolicy: map[string]*botarena.PolicyTotals{
			"assisted":  {Policy: "assisted", Games: 2, Decided: 2, Wins: 1, Losses: 1, WinRate: 0.5, CILow: 0.095, CIHigh: 0.905, Null: 0.5},
			"heuristic": {Policy: "heuristic", Games: 2, Decided: 2, Wins: 1, Losses: 1, WinRate: 0.5, CILow: 0.095, CIHigh: 0.905, Null: 0.5},
		},
	}
	md := sum.Markdown()
	for _, want := range []string{
		"## Arena — 2 games, 2 seats",
		"assisted(izzet-aggro)",
		"qwen3:14b",
		"q4_k_m",
		"### Play",
		"### Funnel",
		"### Latency",
		"### Games",
		"### Seating",
		// seat-games, decided, wins, win% of decided, interval.
		"| assisted | 2 | 2 | 1 | 50.0% | 9.5%–90.5% |",
		// The two denominators are named in the header rather than
		// left for the reader to infer.
		"| policy | seat-games | decided | wins | win% of decided |",
		// Stalled is per seat-game in this table and per game in the
		// header line above it; one page, two meanings, so the column
		// says which.
		"stalled seat-games",
		// And the footnote that says what the denominator is and why
		// the interval is conservative.
		"DECIDED seat-games",
		"negatively correlated",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("Markdown() is missing %q:\n%s", want, md)
		}
	}
	// Every policy gets a row in every policy table, so two runs can
	// be diffed line for line.
	for _, section := range []string{"Play", "Funnel", "Latency"} {
		body, ok := mdSection(md, section)
		if !ok {
			t.Fatalf("no %s section:\n%s", section, md)
		}
		for _, policy := range []string{"| assisted |", "| heuristic |"} {
			if !strings.Contains(body, policy) {
				t.Errorf("the %s table has no %s row:\n%s", section, policy, body)
			}
		}
	}
}

// mdSection returns the body of one "### Name" section.
func mdSection(md, name string) (string, bool) {
	_, rest, ok := strings.Cut(md, "### "+name+"\n")
	if !ok {
		return "", false
	}
	body, _, _ := strings.Cut(rest, "\n### ")
	return body, true
}

// A dropped record means the harvested corpus is missing windows, so
// the report has to say so on its face rather than in a log line
// nobody kept.
func TestMarkdownReportsDroppedRecords(t *testing.T) {
	base := botarena.Summary{
		Config:    botarena.ConfigSummary{Seats: []botarena.SeatSpec{{Tier: tiers.Heuristic}, {Tier: tiers.Heuristic}}, DecisionLog: "/tmp/decisions"},
		PerPolicy: map[string]*botarena.PolicyTotals{"heuristic": {Policy: "heuristic", Games: 2}},
	}
	clean := base
	clean.DecisionLog = botarena.DecisionLogTotals{Records: 500}
	if md := clean.Markdown(); !strings.Contains(md, "nothing dropped") {
		t.Errorf("a complete corpus should say so:\n%s", md)
	}
	holed := base
	holed.DecisionLog = botarena.DecisionLogTotals{Records: 500, Dropped: 12, DroppedQueue: 9, DroppedCap: 3}
	md := holed.Markdown()
	for _, want := range []string{"12 DROPPED", "9 queue", "3 cap", "holes"} {
		if !strings.Contains(md, want) {
			t.Errorf("Markdown() is missing %q:\n%s", want, md)
		}
	}
}

func TestDecisionLogTotalsAdd(t *testing.T) {
	var tot botarena.DecisionLogTotals
	tot.Add(decisionlog.Stats{Records: 10, Dropped: 2, DroppedQueue: 2, Bytes: 100})
	tot.Add(decisionlog.Stats{Records: 5, Dropped: 1, DroppedCap: 1, Bytes: 50})
	if tot.Records != 15 || tot.Dropped != 3 || tot.DroppedQueue != 2 || tot.DroppedCap != 1 || tot.Bytes != 150 {
		t.Errorf("totals = %+v", tot)
	}
}

// Policies orders by evidence, so the report's first row is the one
// with the most games behind it.
func TestSummaryPoliciesOrder(t *testing.T) {
	sum := botarena.Summary{PerPolicy: map[string]*botarena.PolicyTotals{
		"random":    {Games: 1},
		"heuristic": {Games: 9},
		"assisted":  {Games: 9},
	}}
	if got := sum.Policies(); strings.Join(got, ",") != "assisted,heuristic,random" {
		t.Errorf("Policies() = %v", got)
	}
}

// --- configuration refusals -------------------------------------------

func TestValidateRefusesModelTierWithNoClient(t *testing.T) {
	cfg := botarena.Config{
		Seats: []botarena.SeatSpec{{Tier: tiers.Assisted}, {Tier: tiers.Heuristic}},
		Games: 1,
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("an assisted seat with no model client was accepted; it would play the heuristic under a model tier's name")
	}
	if !strings.Contains(err.Error(), "endpoint") {
		t.Errorf("error should say what to do about it, got %v", err)
	}
}

func TestValidateRefusesBadShapes(t *testing.T) {
	cases := []struct {
		name string
		cfg  botarena.Config
	}{
		{"one seat", botarena.Config{Seats: []botarena.SeatSpec{{Tier: tiers.Heuristic}}, Games: 1}},
		{"no games", botarena.Config{Seats: []botarena.SeatSpec{{Tier: tiers.Heuristic}, {Tier: tiers.Heuristic}}, Games: 0}},
		{"unknown tier", botarena.Config{Seats: []botarena.SeatSpec{{Tier: "genius"}, {Tier: tiers.Heuristic}}, Games: 1}},
		{"deck with no index", botarena.Config{Seats: []botarena.SeatSpec{{Tier: tiers.Heuristic, Deck: "izzet-aggro"}, {Tier: tiers.Heuristic}}, Games: 1}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.cfg.Validate(); err == nil {
				t.Fatal("accepted")
			}
		})
	}
}

func TestSeatSpecLabel(t *testing.T) {
	if got := (botarena.SeatSpec{Tier: tiers.Heuristic}).Label(); got != "heuristic" {
		t.Errorf("Label() = %q", got)
	}
	if got := (botarena.SeatSpec{Tier: tiers.Assisted, Name: "assisted-b"}).Label(); got != "assisted-b" {
		t.Errorf("Label() = %q", got)
	}
}

func TestCuratedDeckWithoutIndexIsAnError(t *testing.T) {
	// Not a thinner deck: seating everyone on vanilla bears while the
	// report says "esper-control" would make the measurement a lie.
	if _, err := botarena.CuratedDeck(nil, "esper-control"); err == nil {
		t.Fatal("a curated deck was dealt with no Scryfall index")
	}
}

func TestBattleDeckIsPlayable(t *testing.T) {
	d := botarena.BattleDeck(uuid.Nil)
	if len(d) != 67 {
		t.Errorf("BattleDeck has %d cards, want 67", len(d))
	}
	lands, creatures := 0, 0
	for _, c := range d {
		switch {
		case strings.Contains(c.TypeLine, "Land"):
			lands++
		case strings.Contains(c.TypeLine, "Creature"):
			creatures++
		}
	}
	if lands < 20 || creatures < 20 {
		t.Errorf("BattleDeck is %d lands and %d creatures — the curve the whole-game tests rely on has moved", lands, creatures)
	}
}

// --- whole games (gated) ----------------------------------------------

// Two heuristic seats, one game, and the three things the arena is
// for: it finishes, it tallies both seats under the right policy, and
// it leaves a decision log a tool can read.
func TestArenaPlaysOneHeuristicGame(t *testing.T) {
	requireGameTests(t)
	dir := t.TempDir()
	dl, err := decisionlog.New(decisionlog.Options{Dir: dir})
	if err != nil {
		t.Fatalf("decisionlog.New: %v", err)
	}
	cfg := botarena.Config{
		Seats:       []botarena.SeatSpec{{Tier: tiers.Heuristic}, {Tier: tiers.Heuristic}},
		Games:       1,
		Seed:        101,
		TurnBudget:  40,
		Wall:        3 * time.Minute,
		DecisionLog: dl,
	}
	var got []botarena.GameResult
	sum, err := botarena.Run(context.Background(), cfg, func(r botarena.GameResult) { got = append(got, r) })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("sink saw %d games, want 1", len(got))
	}
	res := got[0]
	t.Logf("seed %d: state=%s turns=%d winner=%d stalled=%v in %s",
		res.Seed, res.State, res.Turns, res.Winner, res.Stalled, res.Elapsed.Round(time.Millisecond))
	if res.Stalled {
		t.Fatalf("the table stalled:\n%s", res.StallDump)
	}
	if res.Turns == 0 {
		t.Error("no turns were played")
	}

	// Two seats of one game is two seat-games of evidence about the
	// heuristic.
	tot := sum.PerPolicy["heuristic"]
	if tot == nil {
		t.Fatalf("no totals for the heuristic; have %v", sum.Policies())
	}
	if tot.Games != 2 {
		t.Errorf("heuristic played %d seat-games, want 2", tot.Games)
	}
	if tot.Wins+tot.Losses+tot.Draws != tot.Games {
		t.Errorf("wins+losses+draws = %d, games = %d", tot.Wins+tot.Losses+tot.Draws, tot.Games)
	}
	if tot.Runner.Decisions == 0 {
		t.Error("no decisions were counted")
	}
	if tot.Decision.Count == 0 {
		t.Error("no decision latencies were sampled")
	}
	if tot.Meter.Windows == 0 {
		t.Error("the Layer A meter saw no windows — the heuristic TIER is a rules.Filter and must")
	}

	// The decision log is the corpus PR 2's position suite is
	// harvested from, so "it was written" is not enough: it has to
	// parse, and the records have to be about this game.
	if res.DecisionLog == "" {
		t.Fatal("no decision log path was recorded")
	}
	records := 0
	if err := decisionlog.Scan(res.DecisionLog, func(rec decisionlog.Record) error {
		records++
		if rec.Game != res.GameID.String() {
			t.Errorf("record names game %s, want %s", rec.Game, res.GameID)
		}
		if rec.V == 0 {
			t.Error("record has no version")
		}
		return nil
	}); err != nil {
		t.Fatalf("Scan(%s): %v", res.DecisionLog, err)
	}
	if records == 0 {
		t.Error("the decision log is empty")
	}
	// The writer is asynchronous and drops rather than blocks a bot
	// seat, so a run has to be able to say whether its corpus is
	// complete. An ordinary two-seat game must not drop anything.
	st := res.DecisionLogStats
	if st == nil {
		t.Fatal("no decision-log stats were reported")
	}
	if st.Dropped != 0 {
		t.Errorf("the decision log dropped %d records (queue %d, cap %d, error %d) — the corpus has holes",
			st.Dropped, st.DroppedQueue, st.DroppedCap, st.DroppedError)
	}
	if st.Records != int64(records) {
		t.Errorf("the writer counted %d records, the file holds %d", st.Records, records)
	}
	if sum.DecisionLog.Records != st.Records {
		t.Errorf("the summary totals %d records, the game wrote %d", sum.DecisionLog.Records, st.Records)
	}
	t.Logf("decision log: %s, %d records, %d bytes, %d dropped", res.DecisionLog, records, st.Bytes, st.Dropped)
}

// Four seats, two policies, rotation on, two games — the shape a real
// baseline run has, with the model tier swapped for `random` so it
// needs no endpoint.
func TestArenaPlaysAMixedTableWithRotation(t *testing.T) {
	requireGameTests(t)
	cfg := botarena.Config{
		Seats: []botarena.SeatSpec{
			{Tier: tiers.Heuristic},
			{Tier: tiers.Random},
			{Tier: tiers.Heuristic},
			{Tier: tiers.Random},
		},
		Games:      2,
		Seed:       202,
		Rotate:     true,
		TurnBudget: 40,
		Wall:       5 * time.Minute,
	}
	sum, err := botarena.Run(context.Background(), cfg, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(sum.Games) != 2 {
		t.Fatalf("played %d games, want 2", len(sum.Games))
	}
	for _, name := range []string{"heuristic", "random"} {
		tot := sum.PerPolicy[name]
		if tot == nil {
			t.Fatalf("no totals for %s; have %v", name, sum.Policies())
		}
		if tot.Games != 4 {
			t.Errorf("%s played %d seat-games, want 4 (2 seats × 2 games)", name, tot.Games)
		}
		if tot.Null != 0.25 {
			t.Errorf("%s null rate is %v at a four-seat table, want 0.25", name, tot.Null)
		}
	}
	// Rotation must actually have moved the chairs.
	seen := map[int]bool{}
	for i := 0; i < 2; i++ {
		for pos, spec := range botarena.Order(4, i, true) {
			if spec == 0 {
				seen[pos] = true
			}
		}
	}
	if len(seen) != 2 {
		t.Errorf("spec 0 sat in %d distinct chairs over 2 games, want 2", len(seen))
	}
	t.Log("\n" + sum.Markdown())
}

// --- seating balance (S3) ---------------------------------------------

// Rotation balances turn order only over a whole number of rotations.
// The default --games 10 on four seats is not one, and a bias of that
// size is the same order as the difference an arena run is trying to
// resolve — so the run has to say so out loud.
func TestChairBalanceWarning(t *testing.T) {
	cases := []struct {
		seats, games int
		rotate, warn bool
	}{
		{4, 10, true, true},   // the default run: 3, 2, 2, 3 first chairs
		{4, 12, true, false},  // a whole number of rotations
		{4, 8, true, false},   //
		{4, 10, false, false}, // no rotation, no rotation bias to warn about
		{2, 3, true, true},
		{2, 4, true, false},
		{3, 10, true, true},
	}
	for _, c := range cases {
		got := botarena.ChairBalanceWarning(c.seats, c.games, c.rotate)
		if (got != "") != c.warn {
			t.Errorf("ChairBalanceWarning(%d, %d, %v) = %q, want warning=%v", c.seats, c.games, c.rotate, got, c.warn)
		}
		if c.warn && !strings.Contains(got, "rotation") {
			t.Errorf("the warning should name the problem, got %q", got)
		}
	}
}

// The histogram is the audit trail: a finished run has to show where
// everybody actually sat, not assert that rotation took care of it.
func TestChairCounts(t *testing.T) {
	got := botarena.ChairCounts(4, 10, true)
	if len(got) != 4 {
		t.Fatalf("ChairCounts gave %d rows, want 4", len(got))
	}
	first := []int{}
	for _, row := range got {
		total := 0
		for _, n := range row {
			total += n
		}
		if total != 10 {
			t.Errorf("a contestant sat in %d of 10 games: %v", total, row)
		}
		first = append(first, row[0])
	}
	// 10 games over 4 chairs: two contestants take chair 0 three
	// times and two take it twice. That IS the residual bias.
	threes, twos := 0, 0
	for _, n := range first {
		switch n {
		case 3:
			threes++
		case 2:
			twos++
		default:
			t.Errorf("chair 0 counts are %v, want threes and twos", first)
		}
	}
	if threes != 2 || twos != 2 {
		t.Errorf("chair 0 counts are %v, want two 3s and two 2s", first)
	}

	// A whole number of rotations is perfectly balanced.
	for _, row := range botarena.ChairCounts(4, 12, true) {
		for pos, n := range row {
			if n != 3 {
				t.Errorf("12 games over 4 seats: chair %d got %d games, want 3 (%v)", pos, n, row)
			}
		}
	}
	// No rotation seats everybody where they were configured.
	for k, row := range botarena.ChairCounts(3, 6, false) {
		if row[k] != 6 {
			t.Errorf("no rotation: contestant %d sat in its own chair %d of 6 games (%v)", k, row[k], row)
		}
	}
}

// --- the win-rate denominator (S2) ------------------------------------

// The win rate is over DECIDED seat-games, because the null rate it is
// compared against is P(win | somebody won). Counting draws in the
// denominator scales every policy down by the decided fraction and
// leaves the null where it is, so a policy beating chance by twenty
// points reports as beating nothing.
func TestBeatsUsesTheDecidedDenominator(t *testing.T) {
	// 20 seat-games at a four-seat table, 8 of them undecided (the
	// turn budget), 6 wins from the 12 that ended: 50% of decided,
	// double the 25% null, and every one of the twelve says so.
	lo, hi := botarena.Wilson(6, 12, botarena.DefaultZ)
	decided := botarena.PolicyTotals{
		Policy: "assisted", Games: 20, Decided: 12, Wins: 6, Losses: 6, Draws: 8,
		WinRate: 0.5, CILow: lo, CIHigh: hi, Null: 0.25,
	}
	if !decided.Beats() {
		t.Errorf("6 wins of 12 decided at a 25%% null should beat the null: CI %.3f–%.3f", lo, hi)
	}
	// The same run read on the old denominator: 6/20 = 30%, whose
	// interval straddles 25%. That is the reading this guards against.
	oldLo, _ := botarena.Wilson(6, 20, botarena.DefaultZ)
	if oldLo > 0.25 {
		t.Fatal("the fixture no longer demonstrates the bug: 6/20 clears the null too")
	}
	// Nothing decided is no evidence, whatever the seat-game count.
	none := botarena.PolicyTotals{Policy: "heuristic", Games: 20, Draws: 20, CILow: 0, CIHigh: 1, Null: 0.25}
	if none.Beats() {
		t.Error("a run that decided nothing must not beat the null")
	}
}

// --- whole-game supervision (fast: these games are cut short) ----------

// A game that runs out of wall clock is a STALL: it is reported, it is
// tallied, and its dump says so. The undecided seat-games land in
// Draws and leave the win rate's denominator empty — no wins from no
// decided games is the whole [0, 1] range, not a confident zero.
func TestWallClockExhaustionIsAStall(t *testing.T) {
	cfg := botarena.Config{
		Seats: []botarena.SeatSpec{{Tier: tiers.Heuristic}, {Tier: tiers.Heuristic}},
		Games: 1,
		Seed:  303,
		// One nanosecond: the deadline is already gone when the
		// supervision loop takes its first look, so this costs a deal
		// and a shuffle rather than a game.
		Wall: time.Nanosecond,
	}
	var got []botarena.GameResult
	sum, err := botarena.Run(context.Background(), cfg, func(r botarena.GameResult) { got = append(got, r) })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("sink saw %d games, want 1", len(got))
	}
	res := got[0]
	if !res.Stalled {
		t.Error("a game that exhausted its wall clock should be marked stalled")
	}
	if res.Aborted {
		t.Error("the run was not cancelled; the game must not be marked aborted")
	}
	if !strings.Contains(res.StallDump, "wall clock") {
		t.Errorf("the dump should name the cause, got %q", res.StallDump)
	}
	if sum.Stalls() != 1 {
		t.Errorf("the summary counts %d stalled games, want 1", sum.Stalls())
	}
	tot := sum.PerPolicy["heuristic"]
	if tot == nil {
		t.Fatalf("no totals for the heuristic; have %v", sum.Policies())
	}
	if tot.Games != 2 || tot.Draws != 2 || tot.Decided != 0 {
		t.Errorf("seat-games=%d draws=%d decided=%d, want 2, 2, 0", tot.Games, tot.Draws, tot.Decided)
	}
	// Stalled is counted per SEAT-game here, and the report's column
	// header says so; Summary.Stalls() counts games.
	if tot.Stalled != 2 {
		t.Errorf("the policy's stalled seat-games = %d, want 2 (both chairs of one stalled game)", tot.Stalled)
	}
	if tot.WinRate != 0 || tot.CIHigh != 1 || tot.CILow != 0 {
		t.Errorf("no decided games is no evidence: rate=%v CI=[%v, %v], want 0 and the whole range", tot.WinRate, tot.CILow, tot.CIHigh)
	}
	if tot.Beats() {
		t.Error("a policy that decided nothing must not beat the null")
	}
}

// A Ctrl-C mid-game is an interruption, not a stall and not a result.
// The partial game must not be labelled "wall clock exhausted" and
// must not be folded into anybody's tallies — a fraction of a game in
// the numerator of a win rate is worse than a missing game.
func TestParentCancellationIsNotAStall(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		time.Sleep(150 * time.Millisecond)
		cancel()
	}()
	cfg := botarena.Config{
		Seats:      []botarena.SeatSpec{{Tier: tiers.Heuristic}, {Tier: tiers.Heuristic}},
		Games:      4,
		Seed:       404,
		TurnBudget: 60,
		Wall:       5 * time.Minute,
	}
	sum, err := botarena.Run(ctx, cfg, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run returned %v, want a cancellation", err)
	}
	// Whatever the cancel interrupted, nothing in this run stalled:
	// the wall was five minutes and the stall threshold fifteen
	// seconds.
	for _, g := range sum.Games {
		if g.Stalled {
			t.Errorf("game seed %d is reported as stalled; a cancelled game is not a stalled one", g.Seed)
		}
	}
	// Every recorded game contributed exactly its seats, and the
	// interrupted one contributed nothing.
	seatGames := 0
	for _, tot := range sum.PerPolicy {
		seatGames += tot.Games
	}
	if want := 2 * len(sum.Games); seatGames != want {
		t.Errorf("%d seat-games tallied over %d completed games, want %d — a partial game was counted",
			seatGames, len(sum.Games), want)
	}
	if sum.Elapsed == 0 {
		t.Error("a cancelled run still has to report how long it ran")
	}
	if sum.PerPolicy == nil {
		t.Error("a cancelled run still has to report the tallies it has")
	}
}

// A game that cannot be STARTED is an error — and the games that were
// already played are still evidence. Returning the empty tally map the
// run started with would print a report with empty tables over a
// games.jsonl full of real games.
func TestRunErrorKeepsTheGamesItPlayed(t *testing.T) {
	dir := t.TempDir()
	logDir := filepath.Join(dir, "decisions")
	dl, err := decisionlog.New(decisionlog.Options{Dir: logDir})
	if err != nil {
		t.Fatalf("decisionlog.New: %v", err)
	}
	cfg := botarena.Config{
		Seats:       []botarena.SeatSpec{{Tier: tiers.Heuristic}, {Tier: tiers.Heuristic}},
		Games:       3,
		Seed:        505,
		Rotate:      true,
		Wall:        time.Nanosecond, // as above: a deal, not a game
		DecisionLog: dl,
	}
	played := 0
	sum, runErr := botarena.Run(context.Background(), cfg, func(r botarena.GameResult) {
		played++
		if played == 1 {
			// Take the decision log's directory away, so the SECOND
			// game cannot be started. A regular file in its place
			// fails for root too, which is who the container's tests
			// run as.
			if err := os.RemoveAll(logDir); err != nil {
				t.Errorf("RemoveAll: %v", err)
			}
			if err := os.WriteFile(logDir, []byte("not a directory"), 0o600); err != nil {
				t.Errorf("WriteFile: %v", err)
			}
		}
	})
	if runErr == nil {
		t.Fatal("a game that could not be started was not reported as an error")
	}
	if played != 1 {
		t.Fatalf("%d games reached the sink, want 1", played)
	}
	if sum.Elapsed == 0 {
		t.Error("the summary has no elapsed time; the report would say `elapsed 0s` over real games")
	}
	tot := sum.PerPolicy["heuristic"]
	if tot == nil {
		t.Fatalf("the tallies of the game that DID play were thrown away; have %v", sum.Policies())
	}
	if tot.Games != 2 {
		t.Errorf("heuristic has %d seat-games, want 2 (both chairs of the one game that played)", tot.Games)
	}
	if len(sum.Games) != 1 {
		t.Errorf("the summary lists %d games, want 1", len(sum.Games))
	}
	if sum.Config.Games != 3 {
		t.Errorf("the summary forgot the requested run length: got %d, want 3", sum.Config.Games)
	}
	if len(sum.Config.Chairs) != len(cfg.Seats) {
		t.Fatalf("the partial summary has %d chair rows, want %d", len(sum.Config.Chairs), len(cfg.Seats))
	}
	for k, row := range sum.Config.Chairs {
		total := 0
		for _, n := range row {
			total += n
		}
		if total != len(sum.Games) {
			t.Errorf("contestant %d's chair histogram counts %d games, but the partial summary contains %d: %v", k, total, len(sum.Games), row)
		}
	}
	if sum.Config.ChairWarning == "" {
		t.Error("one completed game over two rotated seats should report the residual chair imbalance")
	}
	// And the report renders rather than printing empty tables.
	if md := sum.Markdown(); !strings.Contains(md, "| heuristic |") {
		t.Errorf("the report has no row for the games that played:\n%s", md)
	}
}
