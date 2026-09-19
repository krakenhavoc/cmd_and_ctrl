package aiseat_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // catalog hooks
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

const oracleLightningBolt = "4457ed35-7c10-48c8-9776-456485fdf070"

// --- decks --------------------------------------------------------

// monoRedDeck is a small, real-enough deck: the catalog's Lightning
// Bolt, vanilla bears, and Mountains, with a bear as commander.
// Everything in it is something the enumerator can offer and the
// engine can execute, so a random bot playing it exercises lands,
// casts, targets, combat, and damage.
func monoRedDeck(owner uuid.UUID) []game.Card {
	deck := []game.Card{}
	cmdr := game.NewCommander("Commander Bear", owner)
	cmdr.TypeLine = "Legendary Creature — Bear"
	cmdr.ManaCost = "{2}{R}"
	cmdr.Power, cmdr.Toughness = 3, 3
	deck = append(deck, cmdr)
	for i := 0; i < 14; i++ {
		c := game.NewCard("Mountain", owner)
		c.TypeLine = "Basic Land — Mountain"
		deck = append(deck, c)
	}
	for i := 0; i < 10; i++ {
		c := game.NewCard("Bear", owner)
		c.TypeLine = "Creature — Bear"
		c.ManaCost = "{1}{R}"
		c.Power, c.Toughness = 2, 2
		deck = append(deck, c)
	}
	for i := 0; i < 6; i++ {
		c := game.NewCard("Lightning Bolt", owner)
		c.TypeLine = "Instant"
		c.ManaCost = "{R}"
		c.OracleID = oracleLightningBolt
		deck = append(deck, c)
	}
	return deck
}

func newRoom(t *testing.T, seats int, seed uint64) *ws.Room {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < seats; i++ {
		if _, err := g.AddPlayer(fmt.Sprintf("Bot%d", i), monoRedDeck(uuid.Nil)); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(seed, seed+1))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return ws.NewRoom(g, testLogger(), "")
}

func testLogger() *slog.Logger {
	lvl := slog.LevelWarn
	if os.Getenv("AISEAT_DEBUG") != "" {
		lvl = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl}))
}

// --- scripted policy ----------------------------------------------

// scripted picks the first move whose label has one of the given
// prefixes, in order of preference; else passes; else index 0.
type scripted struct {
	prefer []string
	mu     sync.Mutex
	seen   []string
}

func (s *scripted) Name() string { return "scripted" }

func (s *scripted) Decide(_ context.Context, in aiseat.Input) (aiseat.Decision, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.prefer {
		for i, m := range in.Moves {
			if strings.HasPrefix(m.Label, p) {
				s.seen = append(s.seen, m.Label)
				return aiseat.Decision{Index: i, Reason: "scripted"}, nil
			}
		}
	}
	if pi := aiseat.PassIndex(in.Moves); pi >= 0 {
		s.seen = append(s.seen, in.Moves[pi].Label)
		return aiseat.Decision{Index: pi}, nil
	}
	s.seen = append(s.seen, in.Moves[0].Label)
	return aiseat.Decision{Index: 0}, nil
}

type failing struct{ err error }

func (f failing) Name() string { return "failing" }
func (f failing) Decide(context.Context, aiseat.Input) (aiseat.Decision, error) {
	return aiseat.Decision{}, f.err
}

type slow struct{}

func (slow) Name() string { return "slow" }
func (slow) Decide(ctx context.Context, in aiseat.Input) (aiseat.Decision, error) {
	<-ctx.Done()
	return aiseat.Decision{}, ctx.Err()
}

type outOfRange struct{}

func (outOfRange) Name() string { return "oob" }
func (outOfRange) Decide(context.Context, aiseat.Input) (aiseat.Decision, error) {
	return aiseat.Decision{Index: 99}, nil
}

type recordingBroadcaster struct {
	mu    sync.Mutex
	seqs  []uint64
	games map[uuid.UUID]bool
}

func (b *recordingBroadcaster) BroadcastState(gameID uuid.UUID, seq uint64, _ protocol.GameView) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.seqs = append(b.seqs, seq)
	if b.games == nil {
		b.games = map[uuid.UUID]bool{}
	}
	b.games[gameID] = true
}

// handKept reads Player.HandKept under the game's read lock — the
// runners write it concurrently.
//
// PlayerByIDForEffect, not PlayerByID: the lock-free lookup, because
// ReadSnapshot is already holding the read lock. Taking it a second
// time from the same goroutine deadlocks outright whenever a writer
// is queued between the two — Go's RWMutex blocks a new reader behind
// a waiting Lock — and a bot runner dispatching a move is exactly
// that writer. That is #848's first flake: a test that polls this
// helper while two runners play wedges the seat mid-KeepHand and the
// wait times out, which looks for all the world like a slow machine.
// Anything called inside a ReadSnapshot body has to be a *ForEffect /
// *Locked accessor or a plain field read, here as much as in the
// engine.
func handKept(g *game.Game, id uuid.UUID) bool {
	var kept bool
	g.ReadSnapshot(func() {
		if p := g.PlayerByIDForEffect(id); p != nil {
			kept = p.HandKept
		}
	})
	return kept
}

// waitFor polls cond until it holds. It is the ONE wait primitive in
// this package's tests: every "has the bot got there yet" question is
// asked through it, and none of them is asked with a sleep.
//
// waitForBudget is a backstop for a wedged runner, not a timing
// assumption. That distinction is #848: a bot runs on a goroutine the
// test does not schedule, so a budget tight enough to mean anything is
// a budget a loaded machine blows, and a test that fails on how busy
// the machine is tells you about the machine. Nothing here may depend
// on the budget being reached — a test that needs to know a bot has
// STOPPED asks Runner.Idle, and one that needs it gone cancels the
// context and waits on Done.
const waitForBudget = 30 * time.Second

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(waitForBudget)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s after %s", what, waitForBudget)
}

// waitForRunner waits for a runner to exit, through the same backstop
// as every other wait in this package.
//
// It exists so that "the bot stopped playing" is never a literal
// budget. A `select` on `r.Done()` against a `time.After` reads like a
// wait but asserts a deadline: it says the runner must be gone within
// N seconds, which on a loaded machine is a statement about the
// machine. Every one of those this package had was set to a number
// small enough to be a real claim (3s, 5s) and each of them has been
// a CI failure — #635 is the last of them.
func waitForRunner(t *testing.T, what string, r *aiseat.Runner) {
	t.Helper()
	waitFor(t, what, func() bool {
		select {
		case <-r.Done():
			return true
		default:
			return false
		}
	})
}

// botLandsLocked counts the lands a seat controls. Caller must hold
// the game's read lock — the count is only worth anything when it is
// read in the same snapshot as whatever else the assertion is about.
func botLandsLocked(g *game.Game, seat uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == seat && c.IsLand() {
			n++
		}
	}
	return n
}

// turnKey names one player's turn: the round number plus whose turn it
// is inside that round, because Turn.Number counts rounds and
// Turn.ActiveSeat rotates within one.
type turnKey struct{ number, activeSeat int }

// landDrops is a DecisionObserver that tallies the land plays a runner
// APPLIED, keyed by the turn it applied them in.
//
// It is the runner's own report of what it did, and that is the whole
// point of it. Every earlier version of "one land per turn" proved the
// rule by SAMPLING the battlefield whenever the test goroutine next got
// scheduled, which makes the assertion a claim about the sampler. #848
// moved the sample into the same snapshot as the handover it was about,
// which fixed the two-reads-of-a-moving-board half; #634 is what that
// left behind. Both of this test's waits were for states the table
// merely PASSES THROUGH — "the bot controls exactly one land", and
// "the active seat is 1" — and on this machine a full turn cycle is
// about 95ms, so a test goroutine that loses the processor for one
// cycle finds the first condition already false forever (a 30s timeout)
// and reads the second one a cycle late, counting two turns' land drops
// as one turn's ("bot played 2 lands in one turn").
//
// A tally over the decision log cannot miss either: the runner reports
// every window whether or not anybody is looking, so the history is
// complete however the test goroutine is scheduled, and every predicate
// over it is monotone.
type landDrops struct {
	mu     sync.Mutex
	byTurn map[turnKey]int
	total  int
}

func (l *landDrops) Observe(ev aiseat.DecisionEvent) {
	if !ev.Applied || ev.Index < 0 || ev.Index >= len(ev.Input.Moves) {
		return
	}
	if ev.Input.Moves[ev.Index].Kind != legal.KindLand {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.byTurn == nil {
		l.byTurn = map[turnKey]int{}
	}
	l.byTurn[turnKey{ev.Input.View.Turn.Number, ev.Input.View.Turn.ActiveSeat}]++
	l.total++
}

// count is how many land drops the runner has applied, over the whole
// game. Monotone, which is what makes it safe to wait on.
func (l *landDrops) count() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.total
}

// busiest is the turn that took the most land drops, and how many.
func (l *landDrops) busiest() (turnKey, int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	var worst turnKey
	n := 0
	for k, v := range l.byTurn {
		if v > n {
			worst, n = k, v
		}
	}
	return worst, n
}

// turns is how many distinct turns saw a land drop.
func (l *landDrops) turns() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.byTurn)
}

// --- single-runner behaviour ----------------------------------------

func TestRunnerKeepsHandThenPlaysALandAndPasses(t *testing.T) {
	room := newRoom(t, 2, 1)
	g := room.Game
	bot := g.Seats[0]
	human := g.Seats[1]
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bc := &recordingBroadcaster{}
	drops := &landDrops{}
	pol := &scripted{prefer: []string{"Keep hand", "Play Mountain"}}
	r := aiseat.Start(ctx, room, bot.ID, pol, aiseat.Config{Observer: drops}, bc, testLogger())
	// The "human" in seat 1 is a second runner that only ever keeps
	// and passes, so priority actually comes back around.
	h := aiseat.Start(ctx, room, human.ID, &scripted{prefer: []string{"Keep hand"}}, aiseat.Config{}, nil, testLogger())

	waitFor(t, "both seats to keep", func() bool {
		return handKept(g, bot.ID) && handKept(g, human.ID)
	})
	// Seat 0 is active: upkeep → draw → main. The bot passes through
	// upkeep and draw (nothing castable there), then plays a Mountain
	// in main and, having nothing else, passes. Its hand holds a dozen
	// more Mountains, so the claim is that the SECOND one waits for the
	// next turn.
	//
	// Every predicate a test polls has to be monotone, and that is
	// #634. "The bot controls exactly one land" and "the active seat is
	// 1" are both states this table passes through in well under one
	// 95ms turn cycle; a test goroutine that loses the processor for a
	// cycle — which is all a loaded CI runner has to do — finds the
	// first false forever and reads the second one turn late. Waiting
	// for the second land drop to have HAPPENED is monotone, cannot be
	// missed, and is a stronger thing to wait for besides: it is the
	// window in which a bot that ignored its land drop would already
	// have taken two.
	waitFor(t, "the bot to take its second land drop", func() bool {
		return drops.count() >= 2
	})
	if key, n := drops.busiest(); n > 1 {
		t.Errorf("bot played %d lands in one turn (round %d, seat %d)", n, key.number, key.activeSeat)
	}
	if n := drops.turns(); n < 2 {
		t.Errorf("the bot's %d land drops fell in %d turns; the second was not a fresh turn", drops.count(), n)
	}
	// Stop both seats and wait for the goroutines to be gone before
	// reading the counters. Stats and the broadcast log are two reads
	// of a bot that is otherwise still playing, and a move landing
	// between them fails the equality below with "13 broadcasts for 12
	// applied" — a fact about the reads, not about the runner (#848).
	cancel()
	<-r.Done()
	<-h.Done()
	// The board is frozen now, so it can be compared with the decision
	// log exactly. This is what keeps the tally above honest: a
	// per-turn count that disagrees with the lands actually out is a
	// count of something other than land drops.
	var lands int
	g.ReadSnapshot(func() { lands = botLandsLocked(g, bot.ID) })
	if lands != drops.count() {
		t.Errorf("%d of the bot's lands on the battlefield for %d land drops in its decision log", lands, drops.count())
	}
	st := r.Stats()
	if st.Rejected != 0 {
		t.Errorf("rejected moves: %d", st.Rejected)
	}
	if st.Applied == 0 || st.Passes == 0 {
		t.Errorf("expected applied moves and at least one pass, got %+v", st)
	}
	bc.mu.Lock()
	defer bc.mu.Unlock()
	if len(bc.seqs) != int(st.Applied) || !bc.games[g.ID] {
		t.Errorf("every applied move must be broadcast: %d broadcasts for %d applied", len(bc.seqs), st.Applied)
	}
}

func TestRunnerFallsBackWhenPolicyFails(t *testing.T) {
	for name, pol := range map[string]aiseat.Policy{
		"error":        failing{err: errors.New("boom")},
		"timeout":      slow{},
		"out-of-range": outOfRange{},
	} {
		t.Run(name, func(t *testing.T) {
			room := newRoom(t, 2, 2)
			g := room.Game
			bot := g.Seats[0]
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			r := aiseat.Start(ctx, room, bot.ID, pol, aiseat.Config{MaxThink: 30 * time.Millisecond}, nil, testLogger())
			// In the mulligan window there is no pass move, so the
			// fallback is the first legal move — "Keep hand".
			waitFor(t, "fallback keep", func() bool { return handKept(g, bot.ID) })
			st := r.Stats()
			if st.Fallbacks == 0 || st.Decisions != 0 {
				t.Errorf("expected only fallbacks, got %+v", st)
			}
			if st.Rejected != 0 {
				t.Errorf("fallback moves must still be legal: %+v", st)
			}
		})
	}
}

func TestRunnerExitsOnCancelAndOnGameEnd(t *testing.T) {
	room := newRoom(t, 2, 3)
	g := room.Game
	bot := g.Seats[0]

	ctx, cancel := context.WithCancel(context.Background())
	r := aiseat.Start(ctx, room, bot.ID, aiseat.NewRandomPolicy(rand.NewPCG(1, 1)), aiseat.Config{}, nil, testLogger())
	cancel()
	select {
	case <-r.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("runner did not exit on cancel")
	}

	// Game end: the other seat concedes → state ended → the runner
	// wakes on the commit and exits.
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	r2 := aiseat.Start(ctx2, room, bot.ID, aiseat.NewRandomPolicy(rand.NewPCG(2, 2)), aiseat.Config{}, nil, testLogger())
	if _, _, err := room.ApplyExternal(func() error { return g.Concede(g.Seats[1].ID) }); err != nil {
		t.Fatal(err)
	}
	select {
	case <-r2.Done():
	case <-time.After(2 * time.Second):
		t.Fatalf("runner did not exit when the game ended (state %s)", g.CurrentState())
	}
}

func TestRunnerPacesDecisions(t *testing.T) {
	room := newRoom(t, 2, 4)
	g := room.Game
	bot := g.Seats[0]
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	start := time.Now()
	aiseat.Start(ctx, room, bot.ID, &scripted{prefer: []string{"Keep hand"}}, aiseat.Config{MinThink: 150 * time.Millisecond}, nil, testLogger())
	waitFor(t, "paced keep", func() bool { return handKept(g, bot.ID) })
	if el := time.Since(start); el < 140*time.Millisecond {
		t.Errorf("decision landed after %v; MinThink not honoured", el)
	}
}

// envDuration reads a duration from the environment, falling back to
// def when unset or unparseable. Used for the bot-table budgets so a
// loaded CI runner and a local bisect can disagree about how patient
// to be without either one editing the test.
func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

// --- the fuzzer: four random bots ------------------------------------

// TestFourRandomBotsPlay is the S31 exit-criterion smoke test in
// miniature: four random-policy bots at one table must keep the game
// moving — no deadlock, no rejected moves — until it ends or a turn
// budget is spent. Run with AISEAT_DEBUG=1 for the move log.
func TestFourRandomBotsPlay(t *testing.T) {
	requireGameTests(t)
	const turnBudget = 40
	// The stall detector is the real guard here: a deadlocked table
	// never bumps the sequence again, so any threshold catches it and
	// the only question is how long we wait to be sure. The wall clock
	// is a backstop for "moving, but absurdly slowly".
	//
	// Both budgets were tuned on an idle machine, where this test
	// finishes in about 6 seconds. On the shared self-hosted runner
	// under load they are not survivable: with a dozen jobs in flight
	// the same three seeds blew the 60s wall clock at turns 18-25 and
	// failed three unrelated PRs (#418, #431, #434) that had touched
	// nothing near the bot runner. A test that fails on how busy the
	// machine is tells you about the machine.
	//
	// So: generous defaults, overridable for a local bisect. 15s of
	// no sequence movement is still a deadlock by any reasonable
	// reading, and 300s of wall clock is 50x the idle runtime.
	stall := envDuration("AISEAT_STALL", 15*time.Second)
	wallClock := envDuration("AISEAT_WALLCLOCK", 300*time.Second)
	for _, seed := range []uint64{11, 22, 33} {
		t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
			room := newRoom(t, 4, seed)
			g := room.Game
			ctx, cancel := context.WithTimeout(context.Background(), wallClock)
			defer cancel()

			var runners []*aiseat.Runner
			for i, p := range g.Seats {
				pol := aiseat.NewRandomPolicy(rand.NewPCG(seed, uint64(i)))
				runners = append(runners, aiseat.Start(ctx, room, p.ID, pol, aiseat.Config{}, nil, testLogger()))
			}

			lastSeq, lastMove := room.Seq(), time.Now()
			for {
				snap := g.Snapshot()
				if snap.State != game.StateActive || snap.Turn.Number > turnBudget {
					break
				}
				if seq := room.Seq(); seq != lastSeq {
					lastSeq, lastMove = seq, time.Now()
				} else if time.Since(lastMove) > stall {
					t.Fatalf("table stalled at turn %d step %s priority=%d pending=%d\n%s",
						snap.Turn.Number, snap.Turn.Step, snap.Turn.PriorityHolder, len(g.PendingChoices), describeSeats(g))
				}
				if ctx.Err() != nil {
					t.Fatalf("wall clock exhausted at turn %d", snap.Turn.Number)
				}
				time.Sleep(5 * time.Millisecond)
			}
			cancel()
			var total aiseat.Stats
			for _, r := range runners {
				<-r.Done()
				st := r.Stats()
				total.Applied += st.Applied
				total.Rejected += st.Rejected
				total.Fallbacks += st.Fallbacks
				total.Passes += st.Passes
			}
			snap := g.Snapshot()
			t.Logf("seed %d: state=%s turns=%d applied=%d passes=%d rejected=%d fallbacks=%d",
				seed, snap.State, snap.Turn.Number, total.Applied, total.Passes, total.Rejected, total.Fallbacks)
			// The one rejection the engine's combat model permits: a
			// defender's block enumerated during declare_blockers and
			// dispatched after the step wrapped (BlockGrace is 0 here,
			// so the race is live). Anything else is an enumerator bug.
			for _, r := range runners {
				for _, rej := range r.Stats().Rejections {
					if !isStepRace(rej) {
						t.Errorf("enumerator offered a move the engine refused: %s %q: %v", rej.Type, rej.Label, rej.Err)
					}
				}
			}
			if total.Rejected > int64(snap.Turn.Number) {
				t.Errorf("too many step-race rejections for %d turns: %d", snap.Turn.Number, total.Rejected)
			}
			if total.Fallbacks != 0 {
				t.Errorf("random policy should never need the fallback: %d", total.Fallbacks)
			}
			if snap.State == game.StateActive && snap.Turn.Number <= turnBudget {
				t.Errorf("game neither ended nor reached the turn budget: turn %d", snap.Turn.Number)
			}
			if total.Applied < 40 {
				t.Errorf("suspiciously few moves for %d turns: %d", snap.Turn.Number, total.Applied)
			}
		})
	}
}

func describeSeats(g *game.Game) string {
	var b strings.Builder
	type seatLine struct {
		id         uuid.UUID
		name       string
		life, hand int
		elim       bool
	}
	var seats []seatLine
	g.ReadSnapshot(func() {
		for _, p := range g.Seats {
			seats = append(seats, seatLine{p.ID, p.Name, p.Life, p.Hand.Size(), p.Eliminated})
		}
	})
	for _, p := range seats {
		moves := legal.EnumerateFor(g, p.id)
		fmt.Fprintf(&b, "  %s life=%d hand=%d elim=%v moves=%d\n", p.name, p.life, p.hand, p.elim, len(moves))
		for _, m := range moves {
			fmt.Fprintf(&b, "      %s  %s\n", m.Label, string(m.Params))
		}
	}
	g.ReadSnapshot(func() { b.WriteString(describeChoicesLocked(g)) })
	return b.String()
}

func describeChoicesLocked(g *game.Game) string {
	var b strings.Builder
	for _, c := range g.PendingChoices {
		if c == nil {
			continue
		}
		fmt.Fprintf(&b, "  pending %s chooser=%s reason=%q\n", c.Kind, c.Chooser, c.Reason)
		if c.DamageAssignment == nil {
			continue
		}
		f := c.DamageAssignment
		fmt.Fprintf(&b, "  damage prompt: attacker=%s power=%d trample=%v deathtouch=%v blockers=%v\n", f.AttackerID, f.AttackerPower, f.AllowTrample, f.HasDeathtouch, f.BlockerIDs)
		for _, id := range f.BlockerIDs {
			found := false
			for _, card := range g.Battlefield.Cards {
				if card.InstanceID == id {
					found = true
					fmt.Fprintf(&b, "    blocker %s %s %d/%d marked=%d blocking=%s\n", card.Name, id, card.CurrentPower(), card.CurrentToughness(), card.DamageMarked, card.BlockingTarget)
				}
			}
			if !found {
				fmt.Fprintf(&b, "    blocker %s NOT ON BATTLEFIELD\n", id)
			}
		}
	}
	return b.String()
}

// TestActiveBotHoldsPassForBlockers: with a grace configured, the
// active bot does not pass out of declare_blockers while a defender
// still has a legal block; once the defender has blocked (or has
// nothing to block with) the pass goes through promptly.
func TestActiveBotHoldsPassForBlockers(t *testing.T) {
	room := newRoom(t, 2, 9)
	g := room.Game
	bot := g.Seats[0]
	def := g.Seats[1]
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	// Board: bot has an attacker, the defender one untapped blocker.
	atk := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: atk, Name: "Attacker", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: bot.ID, Controller: bot.ID})
	blk := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: blk, Name: "Blocker", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: def.ID, Controller: def.ID})
	for g.Turn.Step != game.StepDeclareAttackers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.DeclareAttacker(atk, def.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	if g.Turn.Step != game.StepDeclareBlockers {
		t.Fatalf("at %s", g.Turn.Step)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	grace := 2 * time.Second
	start := time.Now()
	// Active bot: pass-only policy, with grace.
	aiseat.Start(ctx, room, bot.ID, &scripted{}, aiseat.Config{BlockGrace: grace}, nil, testLogger())
	// It must still be declare_blockers well inside the grace window,
	// because the defender has an undeclared legal block.
	time.Sleep(300 * time.Millisecond)
	if s := g.Snapshot().Turn.Step; s != game.StepDeclareBlockers {
		t.Fatalf("active bot passed out of declare_blockers inside the grace window (now %s)", s)
	}
	// The defender blocks (by hand), then a pass-only runner takes its
	// seat so priority can wrap. The active bot's hold ends as soon as
	// there is nothing left to block with — well before the 2s grace.
	if _, _, err := room.Apply(def.ID, func() error { return g.DeclareBlocker(blk, atk) }); err != nil {
		t.Fatal(err)
	}
	blocked := time.Now()
	aiseat.Start(ctx, room, def.ID, &scripted{}, aiseat.Config{}, nil, testLogger())
	waitFor(t, "step to leave declare_blockers", func() bool {
		return g.Snapshot().Turn.Step != game.StepDeclareBlockers
	})
	if el := time.Since(blocked); el > grace {
		t.Errorf("pass took %v after the block; the hold should end once nothing is left to block", el)
	}
	if el := time.Since(start); el < 300*time.Millisecond {
		t.Errorf("pass landed after %v, before the grace could have mattered", el)
	}
}

// isStepRace recognises the rejections that are races between seats
// rather than enumerator errors: a combat declaration enumerated
// during its step and dispatched after another seat's pass wrapped
// the step; a pass enumerated just before another seat's move
// changed the priority holder (a choice resolving resets priority to
// the active player); and any move enumerated just before the game
// ended.
func isStepRace(rej aiseat.Rejection) bool {
	if errors.Is(rej.Err, game.ErrGameNotActive) {
		return true
	}
	if rej.Type == legal.TypePassPriority && errors.Is(rej.Err, actions.ErrNotPriorityHolder) {
		return true
	}
	return (rej.Type == legal.TypeDeclareBlocker || rej.Type == legal.TypeDeclareAttacker) && errors.Is(rej.Err, game.ErrWrongStep)
}
