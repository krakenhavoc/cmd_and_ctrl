package aiseat_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// life_cost_game_test.go is the whole-game half of #74: a bot handed
// a card that lets it pay life, playing a real game against the real
// engine and the real enumerator, at a life total where the payment
// would kill it.
//
// It is here rather than next door in aiseat/heuristic because the
// thing being checked is not a number in the scorer — the unit tests
// in heuristic/life_cost_test.go cover those. It is the whole chain:
// the ability is registered in the catalog, the enumerator offers it
// (CR 119.4 makes paying down to zero legal, so the engine WOULD take
// it), the cost reaches the policy on legal.Move, and the policy
// declines. Break any link and the bot removes itself from the game.
//
// The opponent is scenery on purpose. Seat 1 passes, keeps, and
// declines everything else, so the only thing in the game that can
// move seat 0's life total is seat 0 — which is what lets the test
// assert on the life total flatly instead of trying to attribute it.

const oracleGriselbrand = "f759d112-76db-4091-a22b-b9f19ab6fa5f"

// --- scenery -------------------------------------------------------

// passPolicy does nothing: it passes priority whenever it can, keeps
// its opening hand, and answers only the windows the engine parks on
// until they are answered (a pending choice, the cleanup discard). It
// never casts, never attacks and never blocks.
type passPolicy struct{}

func (passPolicy) Name() string { return "scenery" }

func (passPolicy) Decide(_ context.Context, in aiseat.Input) (aiseat.Decision, error) {
	// Blocking windows first: the engine refuses to move on until a
	// pending choice is resolved or the cleanup discard is made, so
	// declining those would stall the table rather than do nothing.
	for i := range in.Moves {
		if in.Moves[i].Kind == legal.KindChoice || in.Moves[i].Type == legal.TypeDiscardSelection {
			return aiseat.Decision{Index: i, Reason: "answer the blocking window"}, nil
		}
	}
	for i := range in.Moves {
		if in.Moves[i].Kind == legal.KindPass {
			return aiseat.Decision{Index: i, Reason: "pass"}, nil
		}
	}
	for i := range in.Moves {
		if in.Moves[i].Type == legal.TypeKeepHand {
			return aiseat.Decision{Index: i, Reason: "keep"}, nil
		}
	}
	return aiseat.Decision{Index: aiseat.Decline, Reason: "scenery"}, nil
}

// --- the recorder --------------------------------------------------

// pick is one decision, with the seat's life total as the policy saw
// it. Recording the life AT THE DECISION is the whole point: "the bot
// ended the game alive" is a weaker claim than "the bot never chose a
// move that would have taken it to zero", and the second is the one
// the fix is about.
type pick struct {
	label string
	kind  legal.Kind
	life  int
	cost  int // life component of the chosen move's cost
}

// recordingPolicy wraps the heuristic and keeps what it chose, plus
// whether a life-costing move was ever on the table at all — a test
// that proves the bot declined a move it was never offered proves
// nothing.
//
// It deliberately does NOT forward ShouldConcede. The runner only
// asks a policy that implements aiseat.Conceder, so wrapping it hides
// the concede heuristic, and that is what we want here: a bot that
// scoops at 3 life ends the game before it can demonstrate anything
// about the move it would have made.
type recordingPolicy struct {
	inner *heuristic.Policy

	mu      sync.Mutex
	picks   []pick
	offered int
}

func (p *recordingPolicy) Name() string { return p.inner.Name() }

func (p *recordingPolicy) Decide(ctx context.Context, in aiseat.Input) (aiseat.Decision, error) {
	life := seatLife(in.View, in.Seat)
	costly := 0
	for i := range in.Moves {
		if c := in.Moves[i].Cost; c != nil && c.Life > 0 {
			costly++
		}
	}
	d, err := p.inner.Decide(ctx, in)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.offered += costly
	if err == nil && d.Index >= 0 && d.Index < len(in.Moves) {
		m := in.Moves[d.Index]
		rec := pick{label: m.Label, kind: m.Kind, life: life}
		if m.Cost != nil {
			rec.cost = m.Cost.Life
		}
		p.picks = append(p.picks, rec)
	}
	return d, err
}

func (p *recordingPolicy) snapshot() ([]pick, int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]pick(nil), p.picks...), p.offered
}

func seatLife(v protocol.GameView, seat uuid.UUID) int {
	for i := range v.Seats {
		if v.Seats[i].ID == seat.String() {
			return v.Seats[i].Life
		}
	}
	return 0
}

// --- the fixture ---------------------------------------------------

// griselbrandTable seats the heuristic opposite the scenery, puts a
// Griselbrand on the heuristic's battlefield and sets its life.
//
// Griselbrand is the sharpest case in the catalog: "Pay 7 life: Draw
// seven cards" has no tap and no mana component, so it is activatable
// any number of times in a row, at instant speed, from any total of 7
// or more — including exactly 7.
func griselbrandTable(t *testing.T, seed uint64, life int) (*ws.Room, *recordingPolicy, []aiseat.Policy) {
	t.Helper()
	room := newBattleRoom(t, 2, seed)
	g := room.Game
	bot := g.Seats[0]
	bot.Life = life
	c := game.Card{
		Name: "Griselbrand", TypeLine: "Legendary Creature — Demon",
		ManaCost: "{4}{B}{B}{B}{B}", Power: 7, Toughness: 7,
		OracleID: oracleGriselbrand,
	}
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = bot.ID, bot.ID
	g.Battlefield.PushTop(c)

	rec := &recordingPolicy{inner: heuristic.New()}
	return room, rec, []aiseat.Policy{rec, passPolicy{}}
}

// assertNeverPaidItsLastLife is the invariant both tests hold: every
// life payment the bot chose left it alive.
func assertNeverPaidItsLastLife(t *testing.T, picks []pick, offered int) {
	t.Helper()
	if offered == 0 {
		t.Fatal("the bot was never offered a move with a life cost — the test proved nothing")
	}
	for _, p := range picks {
		if p.cost > 0 && p.life-p.cost < 1 {
			t.Errorf("at %d life the bot chose %q, a %d-life cost: it paid itself to %d",
				p.life, p.label, p.cost, p.life-p.cost)
		}
	}
}

// --- the tests -----------------------------------------------------

// The bug, played out. At seven life the ability is legal, the
// enumerator offers it, the engine would accept it, and taking it is
// the end of the seat.
func TestBotDoesNotPayItsLastLife(t *testing.T) {
	requireGameTests(t)
	room, rec, policies := griselbrandTable(t, 7407, 7)
	res := playGameIn(t, room, 7407, policies, 12, 120*time.Second)
	picks, offered := rec.snapshot()

	assertNeverPaidItsLastLife(t, picks, offered)

	var eliminated bool
	var life int
	room.Game.ReadSnapshot(func() {
		eliminated, life = room.Game.Seats[0].Eliminated, room.Game.Seats[0].Life
	})
	if eliminated || life < 1 {
		t.Errorf("the bot is out of the game at %d life; nothing but its own activation could have done that (eliminated=%v)",
			life, eliminated)
	}
	t.Logf("started at 7 life, finished at %d after %d turns; %d life-costing moves offered, %d decisions recorded",
		life, res.turns, offered, len(picks))
	assertNoEnumeratorBugs(t, res)
}

// The other half, and the reason a hard "never pay life" rule would
// have been a worse fix than the bug: at forty life, drawing seven
// cards for seven of them is one of the best deals in Commander, and
// a bot that will not take it has simply broken differently.
func TestBotStillPaysLifeWhenItIsRight(t *testing.T) {
	requireGameTests(t)
	room, rec, policies := griselbrandTable(t, 7408, 40)
	res := playGameIn(t, room, 7408, policies, 8, 120*time.Second)
	picks, offered := rec.snapshot()

	assertNeverPaidItsLastLife(t, picks, offered)

	paid := 0
	for _, p := range picks {
		if p.kind == legal.KindActivate && p.cost > 0 {
			paid++
		}
	}
	if paid == 0 {
		t.Errorf("the bot never paid seven life from forty in %d turns across %d decisions — "+
			"a policy that refuses every life cost has replaced one bad policy with another",
			res.turns, len(picks))
	}
	var life int
	room.Game.ReadSnapshot(func() { life = room.Game.Seats[0].Life })
	t.Logf("started at 40 life, paid %d times, finished at %d after %d turns", paid, life, res.turns)
	assertNoEnumeratorBugs(t, res)
}
