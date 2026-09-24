package aiseat_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
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
	// board is describeBoard's summary, taken only when the game did
	// NOT end — the one case where the numbers above do not say what
	// happened (#1261).
	board string
	// moves is a lockstep game's move log (#1409); nil for a
	// concurrent one, whose move order is the scheduler's.
	moves []string
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
//
// The seats run concurrently, one runner goroutine each, exactly as a
// production table does — which is what a liveness test wants and
// what makes the game a seed deals different from run to run. A test
// that measures PLAY rather than liveness wants playLockstepGame.
func playGame(t *testing.T, seed uint64, policies []aiseat.Policy, turnBudget int, wall time.Duration) gameResult {
	t.Helper()
	return playGameIn(t, newBattleRoom(t, len(policies), seed), seed, policies, turnBudget, wall)
}

// playLockstepGame is playGame on one goroutine: the seats act in seat
// order, one wake's act-loop each, round after round, so a seed deals
// AND plays one game — the same moves in the same order on every run
// (#1409). Stalls are exact rather than timed: a full round in which
// no seat committed anything is a table nothing will ever move again.
//
// It is the harness for the gates that measure the heuristic's play
// (discussion #1390, question 3: deterministic heuristic quality and
// concurrent liveness are different gates). It does NOT replace the
// concurrent soak, which exists to find the races this one removes.
func playLockstepGame(t *testing.T, seed uint64, policies []aiseat.Policy, turnBudget int, wall time.Duration) gameResult {
	t.Helper()
	return playGameWith(t, newBattleRoom(t, len(policies), seed), seed, policies, turnBudget, wall, true)
}

// playGameIn is playGame against a caller-supplied room, so a caller
// that needs a room built differently — random_game_test.go wants one
// with an on-disk replay log — reuses this loop rather than forking
// it. playGame is the ordinary entry point.
func playGameIn(t *testing.T, room *ws.Room, seed uint64, policies []aiseat.Policy, turnBudget int, wall time.Duration) gameResult {
	t.Helper()
	return playGameWith(t, room, seed, policies, turnBudget, wall, false)
}

// playGameWith is the loop behind both schedules. lockstep false starts
// one runner goroutine per seat and watches the room; lockstep true
// drives stepped runners itself, in seat order, and records the move
// log that makes two runs of a seed comparable.
func playGameWith(t *testing.T, room *ws.Room, seed uint64, policies []aiseat.Policy, turnBudget int, wall time.Duration, lockstep bool) gameResult {
	t.Helper()
	g := room.Game
	if len(g.Seats) != len(policies) {
		t.Fatalf("room has %d seats, got %d policies", len(g.Seats), len(policies))
	}

	// #685: both budgets in this loop were literals, and every
	// whole-game test in the package runs through here. They are now
	// the same two knobs TestFourRandomBotsPlay and TestCatalogSoak
	// already read, so a loaded runner is told to be patient once
	// rather than test by test.
	//
	// The stall detector reads AISEAT_STALL directly: 5s of no
	// sequence movement is a deadlock by any reading, and the only
	// question is how long to wait to be sure. A lockstep game does
	// not need it — see playLockstepGame.
	//
	// The wall clock is a FLOOR, not a replacement. Each caller picks
	// its own — 120s for a 50-turn heuristic table, 25 turns in the
	// decision-log replay — and those numbers say something about the
	// game being played, so the environment may raise them and must
	// never cut them. Unset (the default, and PR CI) changes nothing.
	stall := envDuration("AISEAT_STALL", 5*time.Second)
	if env := envDuration("AISEAT_WALLCLOCK", 0); env > wall {
		wall = env
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
	var observers []aiseat.DecisionObserver
	if gl := openTestDecisionLog(t, g.ID); gl != nil {
		observers = append(observers, gl)
		defer func() {
			if err := gl.Close(); err != nil {
				t.Errorf("close decision log: %v", err)
			}
			t.Logf("decision log: %+v", gl.Stats())
		}()
	}
	var moves *moveLog
	if lockstep {
		moves = newMoveLog(g)
		observers = append(observers, moves)
	}
	switch len(observers) {
	case 0:
	case 1:
		cfg.Observer = observers[0]
	default:
		cfg.Observer = aiseat.ObserverFunc(func(ev aiseat.DecisionEvent) {
			for _, o := range observers {
				o.Observe(ev)
			}
		})
	}
	runners := make([]*aiseat.Runner, 0, len(policies))
	for i, p := range g.Seats {
		if lockstep {
			runners = append(runners, aiseat.NewSteppedRunner(room, p.ID, policies[i], cfg, testLogger()))
		} else {
			runners = append(runners, aiseat.Start(ctx, room, p.ID, policies[i], cfg, nil, testLogger()))
		}
	}

	if lockstep {
		for {
			snap := g.Snapshot()
			if snap.State != game.StateActive || snap.Turn.Round > turnBudget {
				break
			}
			if ctx.Err() != nil {
				t.Fatalf("wall clock exhausted (seed %d) at turn %d", seed, snap.Turn.Round)
			}
			before := room.Seq()
			for _, r := range runners {
				if !r.StepForTest(ctx) {
					break
				}
			}
			if room.Seq() == before && g.CurrentState() == game.StateActive {
				t.Fatalf("table stalled (seed %d, a full lockstep round committed nothing) at turn %d step %s priority=%d pending=%d\n%s",
					seed, snap.Turn.Round, snap.Turn.Step, snap.Turn.PriorityHolder, len(g.PendingChoices), describeSeats(g))
			}
		}
	} else {
		lastSeq, lastMove := room.Seq(), time.Now()
		for {
			snap := g.Snapshot()
			if snap.State != game.StateActive || snap.Turn.Round > turnBudget {
				break
			}
			if seq := room.Seq(); seq != lastSeq {
				lastSeq, lastMove = seq, time.Now()
			} else if time.Since(lastMove) > stall {
				t.Fatalf("table stalled (seed %d, no seq movement in %s) at turn %d step %s priority=%d pending=%d\n%s",
					seed, stall, snap.Turn.Round, snap.Turn.Step, snap.Turn.PriorityHolder, len(g.PendingChoices), describeSeats(g))
			}
			if ctx.Err() != nil {
				t.Fatalf("wall clock exhausted (seed %d) at turn %d", seed, snap.Turn.Round)
			}
			time.Sleep(2 * time.Millisecond)
		}
	}
	cancel()
	for _, r := range runners {
		<-r.Done()
	}

	snap := g.Snapshot()
	res := gameResult{seed: seed, turns: snap.Turn.Round, state: snap.State, winner: -1, elapsed: time.Since(started)}
	if moves != nil {
		res.moves = moves.lines()
	}
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
		// ADR 0057: the engine names the winner, and an effect win
		// leaves several seats standing.
		if o := g.Outcome; o != nil && o.Kind == game.OutcomeWin {
			for i, p := range g.Seats {
				if p.ID == o.Winner {
					res.winner = i
				}
			}
		}
		if res.state != game.StateEnded {
			res.board = describeBoardLocked(g)
		}
	})
	for _, r := range runners {
		res.stats = append(res.stats, r.Stats())
	}
	return res
}

// moveLog records a lockstep game move by move, in a form two runs of
// the same seed can be compared in (#1409): seat index, the room
// sequence the move committed at, the label, the move's parameters and
// the policy's reason. Card instance IDs are fresh UUIDs on every run,
// so raw parameters never compare equal even when the games do; each
// ID is replaced by the order it first appeared in, which is the same
// exactly when the games are. Seat IDs become seat indexes.
//
// The lockstep loop feeds it from one goroutine, but it locks anyway,
// because DecisionObserver's contract says it may not be.
type moveLog struct {
	mu    sync.Mutex
	seats map[string]int
	ids   map[string]string
	out   []string
}

func newMoveLog(g *game.Game) *moveLog {
	l := &moveLog{seats: map[string]int{}, ids: map[string]string{}}
	for i, p := range g.Seats {
		l.seats[p.ID.String()] = i
	}
	return l
}

var uuidPattern = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

func (l *moveLog) Observe(ev aiseat.DecisionEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()
	params := ""
	if ev.Index >= 0 && ev.Index < len(ev.Input.Moves) {
		params = uuidPattern.ReplaceAllStringFunc(string(ev.Input.Moves[ev.Index].Params), func(id string) string {
			if i, ok := l.seats[id]; ok {
				return fmt.Sprintf("seat%d", i)
			}
			if c, ok := l.ids[id]; ok {
				return c
			}
			c := fmt.Sprintf("#%d", len(l.ids))
			l.ids[id] = c
			return c
		})
	}
	l.out = append(l.out, fmt.Sprintf("seat%d seq=%d applied=%v %q %s {%s}",
		l.seats[ev.Seat.String()], ev.Seq, ev.Applied, ev.Label, params, ev.Reason))
}

func (l *moveLog) lines() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.out...)
}

// describeBoardLocked summarises a table the turn budget stopped: per
// seat, life and the cards left in each zone, and the creatures it
// controls tallied by name and size.
//
// #1261: the 2026-09-23 nightly's seed 104 stopped at turn 51 with two
// seats alive, and the artifact said only "lives [-4 6 8 -1]" — not
// whether that was a deadlocked engine, a bot that would not act, or
// two full boards staring at each other. Seeded tables were not
// replayable then (four runner goroutines interleaved differently
// every run), so the artifact was the only look anyone got at that
// board. The locally reproduced one was the third: 22 lands and 22
// untapped creatures a side, libraries at 9, both players under 7
// life — and a lethal swing on the board that lethalPush could not see
// until unblockedPower (#1261). Since #1409 the heuristic gate plays
// lockstep, so the seed in the artifact replays the game exactly; the
// board is still printed because it is the fastest read of a stop.
//
// Caller holds g's read lock.
func describeBoardLocked(g *game.Game) string {
	var b strings.Builder
	for _, p := range g.Seats {
		fmt.Fprintf(&b, "\n  %s life=%d eliminated=%v hand=%d library=%d graveyard=%d:",
			p.Name, p.Life, p.Eliminated, p.Hand.Size(), p.Library.Size(), p.Graveyard.Size())
		tally := map[string]int{}
		var names []string
		lands, tapped := 0, 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller != p.ID {
				continue
			}
			if c.IsLand() {
				lands++
				continue
			}
			if !c.IsCreature() {
				continue
			}
			if c.Tapped {
				tapped++
			}
			key := fmt.Sprintf("%s %d/%d", c.Name, c.CurrentPower(), c.CurrentToughness())
			if tally[key] == 0 {
				names = append(names, key)
			}
			tally[key]++
		}
		sort.Strings(names)
		fmt.Fprintf(&b, " %d lands, %d tapped creatures;", lands, tapped)
		for _, n := range names {
			fmt.Fprintf(&b, " %dx %s", tally[n], n)
		}
	}
	return b.String()
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
	heuristicGateAtSeats(t, 4, 101)
}

// TestTwoHeuristicBotsPlayToAWinner is the same gate on the shape
// #1096 found untested: a HEADS-UP table, which is what ADR 0076
// makes the tutorial's and what #39 fixed the seating for.
//
// A SIBLING rather than a seat count on the test above, because that
// test's name is S31 exit criterion 2 quoted verbatim — "four
// heuristic bots" — and a name that says four while running two is
// worse than a second function. The body is shared; only the seat
// count and the seed block differ.
//
// Its own seed block (201+) so a failure names one table
// unambiguously: reusing 101.. would make "seed 103" mean two
// different games depending on which test printed it.
func TestTwoHeuristicBotsPlayToAWinner(t *testing.T) {
	heuristicGateAtSeats(t, 2, 201)
}

// heuristicGateAtSeats is the gate itself: N seeded tables of `seats`
// heuristic bots, each to a single survivor inside the turn budget,
// with zero engine-refused moves and zero runner fallbacks.
func heuristicGateAtSeats(t *testing.T, seats int, firstSeed uint64) {
	t.Helper()
	requireGameTests(t)
	const (
		turnBudget = 50
		// The default wall clock, not the wall clock. playGameIn takes
		// AISEAT_WALLCLOCK as a floor over this, so the nightly's
		// twenty-game gate can be patient on a contended runner without
		// this number — which is about a 50-turn table, not about the
		// machine — being edited (#685, #600, #606).
		wall = 120 * time.Second
	)
	// Three seeds in the ordinary run; AISEAT_HEURISTIC_GAMES=N for a
	// wider sample when tuning the weights, and 20 on the nightly,
	// which is S31 exit criterion 2's number.
	games := 3
	if n, err := strconv.Atoi(os.Getenv("AISEAT_HEURISTIC_GAMES")); err == nil && n > 0 {
		games = n
	}
	seeds := make([]uint64, 0, games)
	for i := 0; i < games; i++ {
		seeds = append(seeds, firstSeed+uint64(i))
	}
	// #1409: the gate plays lockstep, so a seed is a whole game and a
	// red night replays move for move with the same seed. It measures
	// the heuristic's PLAY; concurrent liveness is the soak's job
	// (AISEAT_SOAK_POLICY=heuristic), and discussion #1390 made them
	// separate gates. AISEAT_HEURISTIC_SCHEDULE=concurrent plays the
	// gate the old way, one goroutine per seat — for comparing the two,
	// not for the nightly.
	schedule := "lockstep"
	if os.Getenv("AISEAT_HEURISTIC_SCHEDULE") == "concurrent" {
		schedule = "concurrent"
	}
	// Logged before the games rather than only per game, so a nightly
	// artifact names the whole sample even when every game passes: the
	// seeds are fixed, so "which twenty" is reproducible from the log
	// alone.
	t.Logf("heuristic gate: %d games at %d seats, seeds %d..%d, turn budget %d, wall %s (AISEAT_WALLCLOCK raises it), %s schedule",
		games, seats, seeds[0], seeds[len(seeds)-1], turnBudget, wall, schedule)
	for _, seed := range seeds {
		t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
			var res gameResult
			if schedule == "concurrent" {
				res = playGame(t, seed, heuristicSeats(seats), turnBudget, wall)
			} else {
				res = playLockstepGame(t, seed, heuristicSeats(seats), turnBudget, wall)
			}
			total := res.totals()
			t.Logf("seed %d: seats=%d state=%s turns=%d winner=%d lives=%v applied=%d passes=%d rejected=%d in %v",
				seed, seats, res.state, res.turns, res.winner, res.lives, total.Applied, total.Passes, total.Rejected, res.elapsed)
			if res.state != game.StateEnded {
				t.Errorf("game did not finish inside %d turns (state %s, lives %v); the board it stopped on:%s",
					turnBudget, res.state, res.lives, res.board)
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
// live in the ordinary test run. Lockstep, like the gate above: it is
// a measurement of play, and a measurement that moves between runs of
// the same seeds is two measurements (#1409).
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
		res := playLockstepGame(t, seed, policies, turnBudget, wall)
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
	waitForRunner(t, "the runner to exit after its policy conceded", r)
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
	// KnownBy the whole table, as a permanent on the battlefield
	// always is. Without it (#95) the bot's FILTERED view showed a
	// face-down card with no type line, `hopeless` counted the
	// opponent's creatures as zero and the seat never conceded — the
	// test passed only because the Killer went on to kill it, which is
	// not what its name says it checks.
	killer := game.Card{
		InstanceID: uuid.New(), Name: "Killer", TypeLine: "Creature — Bear",
		Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
		KnownBy: map[uuid.UUID]bool{bot.ID: true, opp.ID: true},
	}
	g.Battlefield.PushTop(killer)

	// The fixture's own precondition, asserted rather than assumed
	// (#635). `hopeless` requires an EMPTY hand, so a draw landing
	// before the bot's first decision would make the position look
	// survivable and the seat would play on until the budget ran out.
	// It cannot happen here, and this is why: the bot is the active
	// player holding priority in its own UPKEEP, one step before its
	// own draw. Nothing can advance past that step without the bot
	// passing first, so its first window is necessarily a window with
	// an empty hand — whatever order the two runner goroutines happen
	// to be scheduled in. Seat 0 being the active player is a fact
	// about seed 88; if a seeding change ever moves it, this says so
	// instead of timing out.
	g.ReadSnapshot(func() {
		if g.Turn.ActiveSeat != 0 || g.Turn.PriorityHolder != 0 {
			t.Fatalf("fixture: want the bot active and holding priority, got active=%d priority=%d",
				g.Turn.ActiveSeat, g.Turn.PriorityHolder)
		}
		if g.Turn.Step != game.StepUpkeep {
			t.Fatalf("fixture: want the bot on %s, before its draw; got %s", game.StepUpkeep, g.Turn.Step)
		}
		if n := g.Seats[0].Hand.Size(); n != 0 {
			t.Fatalf("fixture: the bot holds %d cards; a hopeless position has none", n)
		}
	})

	cfg := heuristic.DefaultConfig()
	cfg.ConcedeTurns = 1
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := aiseat.Start(ctx, room, bot.ID, heuristic.NewWithConfig(cfg), aiseat.Config{}, nil, testLogger())
	// A pass-only opponent keeps the cursor moving without ever
	// attacking, so the elimination below can only be the concede.
	aiseat.Start(ctx, room, opp.ID, &scripted{}, aiseat.Config{}, nil, testLogger())
	// The runner exiting is the concede: nothing else ends it while
	// the context is live. Waited for through the package's one wait
	// primitive, whose budget is a backstop for a wedged runner and
	// not a claim about how fast a loaded machine decides — the 5s
	// literal this replaces was exactly that claim, and #635 is the
	// CI run that called it (#848's rule, applied here).
	waitForRunner(t, "the bot to concede a hopeless position", r)
	var eliminated bool
	var life int
	g.ReadSnapshot(func() { eliminated, life = g.Seats[0].Eliminated, g.Seats[0].Life })
	if !eliminated {
		t.Fatal("the bot did not concede")
	}
	// It CONCEDED rather than died. The distinction is the test: a
	// seat eliminated at negative life was killed by the Killer, which
	// this fixture would report as a pass while the concede heuristic
	// never ran at all.
	if life <= 0 {
		t.Errorf("the seat is out at %d life — it was killed, not conceded", life)
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
	// "It did not scoop" has to be measured in TURNS, not in seconds:
	// the concede rule counts consecutive hopeless turns, so a wall
	// clock says nothing about how many chances the bot was given. A
	// second and a half of a loaded runner can be no turns at all,
	// which is a test that measured the machine (#635, #848).
	//
	// ConcedeTurns is 1 here, so surviving a third turn is already two
	// turns past the trigger. The scoop check runs on every poll, so
	// the moment it happens this fails with the right message rather
	// than waiting out the bound. The opponent is a real heuristic and
	// may well kill the bot on the way — that ends the game, and a
	// game that ended is as good an answer as turn 3.
	const wantTurn = 3
	waitFor(t, fmt.Sprintf("turn %d, or the game to end, without a scoop", wantTurn), func() bool {
		var conceded, over bool
		var turn int
		g.ReadSnapshot(func() {
			conceded = g.Seats[0].Eliminated && g.Seats[0].Life > 0
			turn, over = g.Turn.Round, g.State != game.StateActive
		})
		if conceded {
			t.Fatalf("the bot scooped a game it still had cards for, on turn %d", turn)
		}
		return over || turn >= wantTurn
	})
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
