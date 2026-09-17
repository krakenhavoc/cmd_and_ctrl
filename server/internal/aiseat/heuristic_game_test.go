package aiseat_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// heuristic_game_test.go runs the heuristic policy against the real
// engine: whole games, four seats and two, measured rather than
// asserted about. The policy's own behaviour is unit-tested next door
// in aiseat/heuristic, where a board can be built by hand in a
// microsecond; what can only be checked here is that the thing plays
// a game of Magic to the end without the engine refusing a single
// move.

// --- harness -------------------------------------------------------

// battleDeck is monoRedDeck with enough cards and enough top end to
// let combat decide a game.
//
// This matters more than it looks. monoRedDeck is 31 cards, so every
// four-bot game on it ends the same way: somebody draws from an empty
// library on about round 25 and the survivors win with 30 life. That
// is a fine fuzzer — it exercises the engine either way — and a
// useless measurement, because it scores mulligan luck rather than
// play. 65 cards with a real curve puts the loss condition back where
// the policy can reach it.
func battleDeck(owner uuid.UUID) []game.Card {
	deck := []game.Card{}
	cmdr := game.NewCommander("Commander Bear", owner)
	cmdr.TypeLine = "Legendary Creature — Bear"
	cmdr.ManaCost = "{2}{R}"
	cmdr.Power, cmdr.Toughness = 3, 3
	deck = append(deck, cmdr)
	add := func(n int, build func() game.Card) {
		for i := 0; i < n; i++ {
			deck = append(deck, build())
		}
	}
	add(24, func() game.Card {
		c := game.NewCard("Mountain", owner)
		c.TypeLine = "Basic Land — Mountain"
		return c
	})
	add(12, func() game.Card {
		c := game.NewCard("Bear", owner)
		c.TypeLine = "Creature — Bear"
		c.ManaCost = "{1}{R}"
		c.Power, c.Toughness = 2, 2
		return c
	})
	add(8, func() game.Card {
		c := game.NewCard("Ogre", owner)
		c.TypeLine = "Creature — Ogre"
		c.ManaCost = "{3}{R}"
		c.Power, c.Toughness = 4, 4
		return c
	})
	// Evasion, and the reason it is here is worth stating. Four
	// competent heuristics on a deck of symmetric vanilla ground
	// creatures produce a perfect board stall: every attack is
	// blocked, every trade is even, no life total moves, and the game
	// is decided by whose library runs out first. That is the correct
	// play of a bad deck rather than a bot that cannot finish — but
	// it measures nothing. Fliers and trample are what every real
	// Commander deck has, and what lets a board advantage convert.
	add(8, func() game.Card {
		c := game.NewCard("Drake", owner)
		c.TypeLine = "Creature — Drake"
		c.ManaCost = "{2}{R}"
		c.Power, c.Toughness = 3, 3
		c.Keywords = []string{"flying"}
		return c
	})
	add(6, func() game.Card {
		c := game.NewCard("Wurm", owner)
		c.TypeLine = "Creature — Wurm"
		c.ManaCost = "{5}{R}"
		c.Power, c.Toughness = 7, 7
		c.Keywords = []string{"trample"}
		return c
	})
	add(8, func() game.Card {
		c := game.NewCard("Lightning Bolt", owner)
		c.TypeLine = "Instant"
		c.ManaCost = "{R}"
		c.OracleID = oracleLightningBolt
		return c
	})
	return deck
}

func newBattleRoom(t *testing.T, seats int, seed uint64) *ws.Room {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < seats; i++ {
		if _, err := g.AddPlayer(fmt.Sprintf("Bot%d", i), battleDeck(uuid.Nil)); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(seed, seed+1))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return ws.NewRoom(g, testLogger(), "")
}

type gameResult struct {
	seed    uint64
	turns   int
	state   game.State
	winner  int // seat index, or -1 if nobody had won when we stopped
	lives   []int
	stats   []aiseat.Stats
	elapsed time.Duration
}

func (r gameResult) totals() aiseat.Stats {
	var t aiseat.Stats
	for _, s := range r.stats {
		t.Decisions += s.Decisions
		t.Applied += s.Applied
		t.Rejected += s.Rejected
		t.Fallbacks += s.Fallbacks
		t.Passes += s.Passes
		t.Rejections = append(t.Rejections, s.Rejections...)
	}
	return t
}

// playGame seats one policy per player and runs to a winner, the turn
// budget, or the wall clock. It fails the test on a stall, which is
// the failure mode that matters: a bot that cannot decide leaves a
// table frozen and nobody to tell about it.
func playGame(t *testing.T, seed uint64, policies []aiseat.Policy, turnBudget int, wall time.Duration) gameResult {
	t.Helper()
	return playGameIn(t, newBattleRoom(t, len(policies), seed), seed, policies, turnBudget, wall)
}

// playGameIn is playGame against a caller-supplied room, so a caller
// that needs a room built differently — random_game_test.go wants one
// with an on-disk replay log — reuses this loop rather than forking
// it. playGame is the ordinary entry point.
func playGameIn(t *testing.T, room *ws.Room, seed uint64, policies []aiseat.Policy, turnBudget int, wall time.Duration) gameResult {
	t.Helper()
	g := room.Game
	if len(g.Seats) != len(policies) {
		t.Fatalf("room has %d seats, got %d policies", len(g.Seats), len(policies))
	}
	ctx, cancel := context.WithTimeout(context.Background(), wall)
	defer cancel()

	started := time.Now()
	// AISEAT_DECISION_LOG turns the per-game decision log on for
	// whole-game tests: the cheapest way to get a real, replayable
	// corpus of windows out of the harness that already plays the
	// games. Unset (the default, and CI) costs nothing — the runner
	// builds no event without an observer.
	cfg := aiseat.Config{}
	if gl := openTestDecisionLog(t, g.ID); gl != nil {
		cfg.Observer = gl
		defer func() {
			if err := gl.Close(); err != nil {
				t.Errorf("close decision log: %v", err)
			}
			t.Logf("decision log: %+v", gl.Stats())
		}()
	}
	runners := make([]*aiseat.Runner, 0, len(policies))
	for i, p := range g.Seats {
		runners = append(runners, aiseat.Start(ctx, room, p.ID, policies[i], cfg, nil, testLogger()))
	}

	lastSeq, lastMove := room.Seq(), time.Now()
	for {
		snap := g.Snapshot()
		if snap.State != game.StateActive || snap.Turn.Number > turnBudget {
			break
		}
		if seq := room.Seq(); seq != lastSeq {
			lastSeq, lastMove = seq, time.Now()
		} else if time.Since(lastMove) > 5*time.Second {
			t.Fatalf("table stalled (seed %d) at turn %d step %s priority=%d pending=%d\n%s",
				seed, snap.Turn.Number, snap.Turn.Step, snap.Turn.PriorityHolder, len(g.PendingChoices), describeSeats(g))
		}
		if ctx.Err() != nil {
			t.Fatalf("wall clock exhausted (seed %d) at turn %d", seed, snap.Turn.Number)
		}
		time.Sleep(2 * time.Millisecond)
	}
	cancel()
	for _, r := range runners {
		<-r.Done()
	}

	snap := g.Snapshot()
	res := gameResult{seed: seed, turns: snap.Turn.Number, state: snap.State, winner: -1, elapsed: time.Since(started)}
	g.ReadSnapshot(func() {
		live := -1
		n := 0
		for i, p := range g.Seats {
			res.lives = append(res.lives, p.Life)
			if !p.Eliminated {
				live, n = i, n+1
			}
		}
		if n == 1 {
			res.winner = live
		}
	})
	for _, r := range runners {
		res.stats = append(res.stats, r.Stats())
	}
	return res
}

// assertNoEnumeratorBugs holds the "zero engine-rejected actions"
// exit criterion, with the one documented exception the engine's
// combat model permits (a declaration enumerated inside its step and
// dispatched after another seat's pass wrapped it).
func assertNoEnumeratorBugs(t *testing.T, res gameResult) {
	t.Helper()
	total := res.totals()
	for _, rej := range total.Rejections {
		if !isStepRace(rej) {
			t.Errorf("seed %d: the engine refused an enumerated move: %s %q: %v", res.seed, rej.Type, rej.Label, rej.Err)
		}
	}
	if total.Fallbacks != 0 {
		t.Errorf("seed %d: the heuristic policy needed the runner's fallback %d times — it is the fallback",
			res.seed, total.Fallbacks)
	}
}

func heuristicSeats(n int) []aiseat.Policy {
	out := make([]aiseat.Policy, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, heuristic.New())
	}
	return out
}

// --- the exit criteria ---------------------------------------------

// Four heuristic bots play to a winner within 50 turns — the S31
// sprint's exit criterion for this sub-PR, verbatim.
func TestFourHeuristicBotsPlayToAWinner(t *testing.T) {
	requireGameTests(t)
	const (
		turnBudget = 50
		wall       = 120 * time.Second
	)
	// Three seeds in the ordinary run; AISEAT_HEURISTIC_GAMES=N for a
	// wider sample when tuning the weights.
	games := 3
	if n, err := strconv.Atoi(os.Getenv("AISEAT_HEURISTIC_GAMES")); err == nil && n > 0 {
		games = n
	}
	seeds := make([]uint64, 0, games)
	for i := 0; i < games; i++ {
		seeds = append(seeds, uint64(101+i))
	}
	for _, seed := range seeds {
		t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
			res := playGame(t, seed, heuristicSeats(4), turnBudget, wall)
			total := res.totals()
			t.Logf("seed %d: state=%s turns=%d winner=%d lives=%v applied=%d passes=%d rejected=%d in %v",
				seed, res.state, res.turns, res.winner, res.lives, total.Applied, total.Passes, total.Rejected, res.elapsed)
			if res.state != game.StateEnded {
				t.Errorf("game did not finish inside %d turns (state %s, lives %v)", turnBudget, res.state, res.lives)
			}
			if res.winner < 0 {
				t.Errorf("game ended with no single survivor: lives %v", res.lives)
			}
			assertNoEnumeratorBugs(t, res)
		})
	}
}

// The measurement ADR 0033 asks for: random is "the baseline every
// other policy is measured against", so the heuristic has to beat it
// and the margin has to be recorded rather than assumed.
//
// Heads-up, alternating seats so the turn-order advantage cancels.
// AISEAT_H2H_GAMES raises the sample; the default is small enough to
// live in the ordinary test run.
func TestHeuristicBeatsRandomHeadToHead(t *testing.T) {
	requireGameTests(t)
	games := 8
	if n, err := strconv.Atoi(os.Getenv("AISEAT_H2H_GAMES")); err == nil && n > 0 {
		games = n
	}
	const (
		turnBudget = 60
		wall       = 120 * time.Second
	)
	var wins, losses, draws int
	for i := 0; i < games; i++ {
		seed := uint64(9000 + i)
		heuristicSeat := i % 2
		policies := make([]aiseat.Policy, 2)
		policies[heuristicSeat] = heuristic.New()
		policies[1-heuristicSeat] = aiseat.NewRandomPolicy(rand.NewPCG(seed, 7))
		res := playGame(t, seed, policies, turnBudget, wall)
		switch {
		case res.winner == heuristicSeat:
			wins++
		case res.winner >= 0:
			losses++
		default:
			draws++
		}
		assertNoEnumeratorBugs(t, res)
	}
	decided := wins + losses
	rate := 0.0
	if decided > 0 {
		rate = float64(wins) / float64(decided)
	}
	t.Logf("heuristic vs random over %d games: %d wins, %d losses, %d unfinished — %.0f%% of decided games",
		games, wins, losses, draws, rate*100)
	if decided == 0 {
		t.Fatal("no game reached a winner; the head-to-head measured nothing")
	}
	if rate < 0.7 {
		t.Errorf("heuristic won %.0f%% of decided games against random; a policy that does not beat the baseline convincingly is not done", rate*100)
	}
}

// ADR 0033 §10: MaxThink is a hard deadline and the table never waits
// on a bot. The heuristic is supposed to be free — sub-millisecond —
// and this is the number that says whether it still is.
func TestHeuristicDecisionLatency(t *testing.T) {
	requireGameTests(t)
	timed := make([]*timingPolicy, 4)
	policies := make([]aiseat.Policy, 4)
	for i := range timed {
		timed[i] = &timingPolicy{inner: heuristic.New()}
		policies[i] = timed[i]
	}
	res := playGame(t, 4242, policies, 40, 120*time.Second)
	var all []time.Duration
	for _, p := range timed {
		all = append(all, p.samples()...)
	}
	if len(all) < 200 {
		t.Fatalf("only %d decisions sampled over %d turns; not enough to say anything", len(all), res.turns)
	}
	sort.Slice(all, func(i, j int) bool { return all[i] < all[j] })
	p50 := all[len(all)*50/100]
	p99 := all[len(all)*99/100]
	t.Logf("heuristic decision latency over %d decisions: p50=%v p99=%v max=%v", len(all), p50, p99, all[len(all)-1])
	if p99 > 50*time.Millisecond {
		t.Errorf("p99 decision latency %v — the heuristic tier is meant to be free", p99)
	}
}

// --- the concede path ----------------------------------------------

// alwaysConcedes is the wiring test's stand-in: it passes when asked
// to decide and scoops when asked whether it is done.
type alwaysConcedes struct{}

func (alwaysConcedes) Name() string { return "always-concedes" }

func (alwaysConcedes) Decide(_ context.Context, in aiseat.Input) (aiseat.Decision, error) {
	if pi := aiseat.PassIndex(in.Moves); pi >= 0 {
		return aiseat.Decision{Index: pi}, nil
	}
	return aiseat.Decision{Index: 0}, nil
}

func (alwaysConcedes) ShouldConcede(aiseat.Input) bool { return true }

// Conceding is not a legal MOVE — `legal` deliberately does not
// enumerate it, or the random policy would scoop out of its own
// fuzzer — so it reaches the engine through the optional Conceder
// interface instead. This is that path end to end.
func TestRunnerConcedesThroughTheConcederInterface(t *testing.T) {
	room := newRoom(t, 2, 77)
	g := room.Game
	bot := g.Seats[0]
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := aiseat.Start(ctx, room, bot.ID, alwaysConcedes{}, aiseat.Config{}, nil, testLogger())
	select {
	case <-r.Done():
	case <-time.After(3 * time.Second):
		t.Fatal("the runner did not exit after its policy conceded")
	}
	var eliminated bool
	g.ReadSnapshot(func() { eliminated = g.Seats[0].Eliminated })
	if !eliminated {
		t.Fatal("the seat is still in the game")
	}
}

// And the heuristic's own judgement, in a real game: a seat at 1 life
// with nothing in hand, nothing on the board and an opponent holding
// a creature scoops. ConcedeTurns is dropped to 1 here because the
// three-turn patience is unit-tested in the policy package and
// waiting three real turns for it buys nothing.
func TestHeuristicConcedesAHopelessSeat(t *testing.T) {
	room := newRoom(t, 2, 88)
	g := room.Game
	bot, opp := g.Seats[0], g.Seats[1]
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	// Set the board up by hand, before any runner is watching.
	g.Seats[0].Life = 1
	g.Seats[0].Hand.Cards = nil
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Killer", TypeLine: "Creature — Bear",
		Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})

	cfg := heuristic.DefaultConfig()
	cfg.ConcedeTurns = 1
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := aiseat.Start(ctx, room, bot.ID, heuristic.NewWithConfig(cfg), aiseat.Config{}, nil, testLogger())
	// A pass-only opponent keeps the cursor moving without ever
	// attacking, so the elimination below can only be the concede.
	aiseat.Start(ctx, room, opp.ID, &scripted{}, aiseat.Config{}, nil, testLogger())
	select {
	case <-r.Done():
	case <-time.After(5 * time.Second):
		t.Fatalf("the bot played on from a hopeless position (life %d)", g.Snapshot().Seats[0].Life)
	}
	var eliminated bool
	g.ReadSnapshot(func() { eliminated = g.Seats[0].Eliminated })
	if !eliminated {
		t.Fatal("the bot did not concede")
	}
}

// The other half of "deliberately conservative": a bot that is merely
// LOSING keeps playing. This is the regression that matters — a
// concede heuristic tuned one notch too eager takes a human's win
// away from them.
func TestHeuristicDoesNotConcedeAWinnableGame(t *testing.T) {
	requireGameTests(t)
	room := newRoom(t, 2, 99)
	g := room.Game
	bot, opp := g.Seats[0], g.Seats[1]
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	g.Seats[0].Life = 2 // one point off dead, and still holding a hand
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Killer", TypeLine: "Creature — Bear",
		Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})
	cfg := heuristic.DefaultConfig()
	cfg.ConcedeTurns = 1
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	aiseat.Start(ctx, room, bot.ID, heuristic.NewWithConfig(cfg), aiseat.Config{}, nil, testLogger())
	aiseat.Start(ctx, room, opp.ID, heuristic.New(), aiseat.Config{}, nil, testLogger())
	deadline := time.Now().Add(1500 * time.Millisecond)
	for time.Now().Before(deadline) {
		var conceded bool
		g.ReadSnapshot(func() { conceded = g.Seats[0].Eliminated && g.Seats[0].Life > 0 })
		if conceded {
			t.Fatal("the bot scooped a game it still had cards for")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// timingPolicy records how long each decision took. It is a Policy
// decorator, which is the shape ADR 0033's tiers are built from.
type timingPolicy struct {
	inner aiseat.Policy
	mu    sync.Mutex
	took  []time.Duration
}

func (p *timingPolicy) Name() string { return p.inner.Name() }

func (p *timingPolicy) Decide(ctx context.Context, in aiseat.Input) (aiseat.Decision, error) {
	start := time.Now()
	d, err := p.inner.Decide(ctx, in)
	el := time.Since(start)
	p.mu.Lock()
	p.took = append(p.took, el)
	p.mu.Unlock()
	return d, err
}

func (p *timingPolicy) ShouldConcede(in aiseat.Input) bool {
	c, ok := p.inner.(aiseat.Conceder)
	return ok && c.ShouldConcede(in)
}

func (p *timingPolicy) samples() []time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]time.Duration(nil), p.took...)
}
