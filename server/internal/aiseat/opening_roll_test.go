package aiseat_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// opening_roll_test.go — ADR 0121 §4 and §9: bots in the opening roll.
// Layer A and the heuristic roll and choose themselves, and agree; the
// random tier answers; and bots seated at a table that opens the roll
// roll, choose and play on with nobody's help.

// Key (22, 2048) at four seats: round 1 is 7, 4, 8, 8, so seats 2 and 3
// roll again, and seat 3 wins round 2 with 18 to seat 2's 9. A tie is
// the case worth driving: a seat out of round 2 must sit it out, and
// the tied seats must roll a second time.
const (
	openingSeed1, openingSeed2 = 22, 2048
	openingWinner              = 3
)

func openingRollRoom(t *testing.T) *ws.Room {
	t.Helper()
	g := game.NewGame()
	for i := range 4 {
		if _, err := g.AddPlayer(fmt.Sprintf("Bot%d", i), monoRedDeck(uuid.Nil)); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithOpeningRoll(rand.New(rand.NewPCG(openingSeed1, openingSeed2))); err != nil {
		t.Fatalf("StartWithOpeningRoll: %v", err)
	}
	return ws.NewRoom(g, testLogger(), "")
}

// Every opening-roll window of a real table, decided by Layer A, the
// heuristic and the random tier from the seat's own filtered view: the
// first two take the same move, always the die or "I go first", and
// the third picks some offered move.
func TestPoliciesAnswerEveryOpeningRollWindow(t *testing.T) {
	g := openingRollRoom(t).Game
	h := heuristic.New()
	random := aiseat.NewRandomPolicy(rand.NewPCG(1, 2))
	ctx := context.Background()

	windows, choices := 0, 0
	for g.OpeningRollOpen() {
		acted := false
		for _, p := range g.Seats {
			moves := legal.EnumerateFor(g, p.ID)
			if len(moves) == 0 {
				continue
			}
			in := aiseat.Input{View: protocol.ViewOfGameFor(g, p.ID.String()), Seat: p.ID, Moves: moves}
			v := rules.Resolve(in)
			wantRule := rules.RuleOpeningRoll
			if v.Index >= 0 && v.Index < len(moves) && moves[v.Index].Type == legal.TypeChooseStartingPlayer {
				wantRule = rules.RuleOpeningChoice
			}
			if !v.Absorbed() || v.Rule != wantRule {
				t.Fatalf("seat %d: Layer A did not absorb %+v: %#v", p.Seat, moves, v)
			}
			d, err := h.Decide(ctx, in)
			if err != nil || d.Index != v.Index {
				t.Fatalf("seat %d: heuristic %d (%v) disagrees with Layer A %d", p.Seat, d.Index, err, v.Index)
			}
			r, err := random.Decide(ctx, in)
			if err != nil || r.Index < 0 || r.Index >= len(moves) {
				t.Fatalf("seat %d: random answered %d (%v) over %d moves", p.Seat, r.Index, err, len(moves))
			}
			mv := moves[v.Index]
			switch mv.Type {
			case legal.TypeRollOpening:
			case legal.TypeChooseStartingPlayer:
				choices++
				if mv.Label != "I go first" {
					t.Fatalf("seat %d chose %q, want itself", p.Seat, mv.Label)
				}
			default:
				t.Fatalf("seat %d took %s during the opening roll", p.Seat, mv.Type)
			}
			if err := actions.Dispatch(g, actions.Action{Type: actions.Type(mv.Type), Player: mv.Player, Caller: p.ID, Params: mv.Params}); err != nil {
				t.Fatalf("seat %d: %s refused: %v", p.Seat, mv.Label, err)
			}
			windows++
			acted = true
			break // the board moved; enumerate again from the top
		}
		if !acted {
			t.Fatal("the opening roll is open and nobody has a move")
		}
	}
	// Four dice, two rerolls, one choice.
	if windows != 7 || choices != 1 {
		t.Fatalf("%d windows with %d choices, want 7 and 1", windows, choices)
	}
	if g.StartingSeat != openingWinner {
		t.Fatalf("starting seat %d, want the winner, seat %d, who chose itself", g.StartingSeat, openingWinner)
	}
}

// The random tier stays uniform (ADR 0121 §4): over the winner's
// choice it picks every seat, so a random bot may hand the first turn
// away.
func TestRandomTierChoosesAnySeatToGoFirst(t *testing.T) {
	g := openingRollRoom(t).Game
	for g.OpeningRoll.Chooser < 0 {
		round := g.OpeningRoll.Rounds[len(g.OpeningRoll.Rounds)-1]
		for _, seat := range round.Seats {
			if err := g.RollOpening(g.Seats[seat].ID); err != nil {
				t.Fatal(err)
			}
		}
	}
	winner := g.Seats[g.OpeningRoll.Chooser].ID
	in := aiseat.Input{Seat: winner, Moves: legal.EnumerateFor(g, winner)}
	if len(in.Moves) != 4 {
		t.Fatalf("the winner is offered %+v, want one choice per seat", in.Moves)
	}
	random := aiseat.NewRandomPolicy(rand.NewPCG(7, 11))
	seen := map[int]bool{}
	for range 200 {
		d, err := random.Decide(context.Background(), in)
		if err != nil {
			t.Fatal(err)
		}
		seen[d.Index] = true
	}
	if len(seen) != len(in.Moves) {
		t.Fatalf("the random tier picked only %v of %d choices in 200 tries", seen, len(in.Moves))
	}
}

// openingRollTally is a DecisionObserver over the runners' own reports
// of what they applied (#634: a complete history, not a sample), so
// every count it gives is monotone and safe to wait on.
type openingRollTally struct {
	mu       sync.Mutex
	rolls    map[uuid.UUID]int
	choices  map[uuid.UUID][]string
	notA     []string
	rejected []string
}

func (o *openingRollTally) Observe(ev aiseat.DecisionEvent) {
	if ev.Index < 0 || ev.Index >= len(ev.Input.Moves) {
		return
	}
	mv := ev.Input.Moves[ev.Index]
	if mv.Kind != legal.KindOpeningRoll {
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if !ev.Applied {
		o.rejected = append(o.rejected, fmt.Sprintf("%s: %s: %v", ev.Policy, mv.Label, ev.RejectErr))
		return
	}
	if strings.HasPrefix(ev.Policy, "rules+") && (ev.Trace.Layer != "A" || (ev.Trace.Rule != rules.RuleOpeningRoll && ev.Trace.Rule != rules.RuleOpeningChoice)) {
		o.notA = append(o.notA, fmt.Sprintf("%s: %s (layer %q rule %q)", ev.Policy, mv.Label, ev.Trace.Layer, ev.Trace.Rule))
	}
	if o.rolls == nil {
		o.rolls, o.choices = map[uuid.UUID]int{}, map[uuid.UUID][]string{}
	}
	switch mv.Type {
	case legal.TypeRollOpening:
		o.rolls[ev.Seat]++
	case legal.TypeChooseStartingPlayer:
		o.choices[ev.Seat] = append(o.choices[ev.Seat], mv.Label)
	}
}

// ADR 0121 §4 and §9: bots seated at a table started with
// StartWithOpeningRoll roll, reroll a tie, choose and play on, with
// nobody pressing anything for them. The lobby still starts tables
// with the automatic roll until ADR 0121 PR 5, so this is the
// whole-table check of the window until TestManagerPlaysALobbySeated-
// Table passes through it too.
func TestBotsRollAndChooseThroughTheOpeningRollAndPlayOn(t *testing.T) {
	room := openingRollRoom(t)
	g := room.Game
	tally := &openingRollTally{}
	cfg := aiseat.Config{Observer: tally}
	meter := &rules.Meter{}
	// The heuristic tier (Layer A over the heuristic) at the winner's
	// seat and one tied seat, the bare heuristic at another, and a
	// random seat out of the tie.
	policies := []aiseat.Policy{
		aiseat.NewRandomPolicy(rand.NewPCG(3, 5)),
		heuristic.New(),
		rules.NewFilter(heuristic.New(), meter),
		rules.NewFilter(heuristic.New(), meter),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runners := make([]*aiseat.Runner, 0, len(policies))
	for i, p := range g.Seats {
		runners = append(runners, aiseat.Start(ctx, room, p.ID, policies[i], cfg, nil, testLogger()))
	}

	// Every predicate below is monotone: the window only ever closes,
	// the mulligan only ever closes, and turns only ever begin.
	waitFor(t, "the bots to finish the opening roll and the mulligan and reach turn 3", func() bool {
		done := false
		g.ReadSnapshot(func() {
			done = g.State != game.StateActive ||
				(g.OpeningRoll == nil && !g.MulligansOpen && g.Turn.Seq >= 3)
		})
		return done
	})
	cancel()
	for _, r := range runners {
		waitForRunner(t, "the runner to exit", r)
	}

	tally.mu.Lock()
	defer tally.mu.Unlock()
	if len(tally.rejected) != 0 {
		t.Errorf("opening-roll moves rejected: %v", tally.rejected)
	}
	if len(tally.notA) != 0 {
		t.Errorf("opening-roll windows not absorbed by Layer A: %v", tally.notA)
	}
	for seat, want := range []int{1, 1, 2, 2} {
		if got := tally.rolls[g.Seats[seat].ID]; got != want {
			t.Errorf("seat %d rolled %d dice, want %d", seat, got, want)
		}
	}
	winner := g.Seats[openingWinner].ID
	if got := tally.choices[winner]; len(got) != 1 || got[0] != "I go first" {
		t.Errorf("the winner's choices: %v, want [I go first]", got)
	}
	if len(tally.choices) != 1 {
		t.Errorf("choices by %d seats, want the winner's only: %v", len(tally.choices), tally.choices)
	}
	if st := meter.Stats(); st.ByRule[rules.RuleOpeningRoll] != 4 || st.ByRule[rules.RuleOpeningChoice] != 1 {
		t.Errorf("Layer A absorbed %d opening dice and %d choices, want 4 (two seats' dice, one reroll each) and 1: %+v",
			st.ByRule[rules.RuleOpeningRoll], st.ByRule[rules.RuleOpeningChoice], st.ByRule)
	}
	var starting int
	g.ReadSnapshot(func() { starting = g.StartingSeat })
	if starting != openingWinner {
		t.Errorf("starting seat %d, want the winner, seat %d", starting, openingWinner)
	}
}

// ADR 0121 §4, pacing: a bot rolls on its normal clock, the table's
// bot speed, like any other decision. At the default "normal" pace that
// is at least 700ms (defaultMinThink) after the window opens, so a die
// never lands before a person at the table has seen the Roll button.
// Only a lower bound is asserted, which a slow machine can only make
// truer (#876).
func TestBotsRollOnTheTablesBotPace(t *testing.T) {
	room := openingRollRoom(t)
	g := room.Game
	const normalMinThink = 700 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := time.Now()
	r := aiseat.Start(ctx, room, g.Seats[0].ID, heuristic.New(),
		aiseat.Config{FollowTablePace: true, MaxThink: 2 * time.Second}, nil, testLogger())
	var rolledAt time.Time
	// Monotone: a die rolled in round 1 stays rolled.
	waitFor(t, "seat 0 to roll its die", func() bool {
		rolled := false
		g.ReadSnapshot(func() {
			for _, d := range g.OpeningRoll.Rounds[0].Rolls {
				rolled = rolled || d.Seat == 0
			}
		})
		if rolled && rolledAt.IsZero() {
			rolledAt = time.Now()
		}
		return rolled
	})
	if elapsed := rolledAt.Sub(started); elapsed < normalMinThink {
		t.Errorf("seat 0 rolled %v after the window opened, before the normal pace's %v", elapsed, normalMinThink)
	}
	cancel()
	waitForRunner(t, "the runner to exit", r)
}

// The same, played to a winner: four heuristic-tier seats on the
// production schedule (one runner goroutine each) open the roll
// themselves, roll, choose and finish the game. Gated like every
// whole-game test (AISEAT_GAME_TESTS).
func TestHeuristicTablePlaysToAWinnerThroughTheOpeningRoll(t *testing.T) {
	requireGameTests(t)
	const seed = 121
	g := game.NewGame()
	for i := range 4 {
		if _, err := g.AddPlayer(fmt.Sprintf("Bot%d", i), battleDeck(uuid.Nil)); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithOpeningRoll(rand.New(rand.NewPCG(seed, seed+1))); err != nil {
		t.Fatalf("StartWithOpeningRoll: %v", err)
	}
	if !g.OpeningRollOpen() {
		t.Fatal("the table did not open the opening roll")
	}
	policies := make([]aiseat.Policy, 0, 4)
	for range 4 {
		policies = append(policies, rules.NewFilter(heuristic.New(), nil))
	}
	res := playGameIn(t, ws.NewRoom(g, testLogger(), ""), seed, policies, 50, 120*time.Second)
	total := res.totals()
	t.Logf("seed %d: state=%s turns=%d winner=%d lives=%v applied=%d rejected=%d in %v",
		seed, res.state, res.turns, res.winner, res.lives, total.Applied, total.Rejected, res.elapsed)
	if g.OpeningRollOpen() {
		t.Fatal("the opening roll is still open")
	}
	if res.state != game.StateEnded || res.winner < 0 {
		t.Errorf("game did not finish with a winner (state %s, lives %v):%s", res.state, res.lives, res.board)
	}
	assertNoEnumeratorBugs(t, res)
}
