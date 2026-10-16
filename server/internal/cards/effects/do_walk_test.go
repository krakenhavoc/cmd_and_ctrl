package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// do_walk_test.go — #4012 (CR 608.2c): a Do's later steps wait for the
// answer to a prompt an earlier step asked, and a table waiting on that
// prompt keeps the later steps across a restore and an undo.

const doWalkProbeOracle = "test-4012-do-walk-probe"

// doWalkProbeDamage is "this deals N damage to the first target", a step
// no catalog primitive spells as a value (DealDamage takes fixed IDs).
type doWalkProbeDamage struct{ N int }

func (s doWalkProbeDamage) Apply(ctx *Context) error {
	if len(ctx.Item.Targets) == 0 {
		return nil
	}
	return ctx.Game.DealDamageToCreatureForEffect(ctx.Item.SourceCardID, ctx.Item.Targets[0].ID, s.N)
}

// doWalkAlways is a condition that holds, so an If's steps run.
type doWalkAlways struct{}

func (doWalkAlways) Holds(*Context) bool                   { return true }
func (doWalkAlways) condition() (game.FootprintStep, bool) { return game.FootprintStep{}, false }

func init() {
	Register(Spec{
		OracleID: doWalkProbeOracle,
		Name:     "Do Walk Probe",
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Do Walk Probe deals 1 damage to target creature. Draw a card.",
				Cost:    TapCost(),
				Targets: TargetCreature("target creature"),
				Effect:  Do(doWalkProbeDamage{N: 1}, DrawCards{N: 1}),
			},
			{
				Label:   "{T}: Scry 2, then draw a card.",
				Cost:    TapCost(),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  Do(Scry{N: 2}, DrawCards{N: 1}),
			},
			{
				Label:   "{T}: If it holds, scry 1, then draw a card, then gain 1 life.",
				Cost:    TapCost(),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  Do(If{Cond: doWalkAlways{}, Then: []Applier{Scry{N: 1}, DrawCards{N: 1}}}, GainLife{Amount: 1}),
			},
		},
	})
}

func doWalkProbe(t *testing.T, g *game.Game, me uuid.UUID, idx int, target uuid.UUID) {
	t.Helper()
	probe := pushCatalogPermanent(g, me, "Do Walk Probe", "Artifact", doWalkProbeOracle, false)
	if err := activateOn(t, g, me, probe, idx, target); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
}

// Khârn the Betrayer is dealt damage by the first step of a Do; the
// prevention follow-up asks its controller for an opponent (#2040).
// "Draw a card" is the next instruction, so it waits for the answer —
// on the live table and on one restored while the question is open.
func TestDoStepAfterKharnsQuestionWaitsForTheAnswer(t *testing.T) {
	for _, restored := range []bool{false, true} {
		g := newCatalogGame(t)
		me, third := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
		kharn := kharnPush(g, me.ID, me.ID)
		hand := me.Hand.Size()
		doWalkProbe(t, g, me.ID, 0, kharn)
		if latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil {
			t.Fatal("Khârn's question is not open")
		}
		if me.Hand.Size() != hand {
			t.Fatalf("hand %d with the question open, want %d: the draw ran before the answer", me.Hand.Size(), hand)
		}
		if restored {
			g = restoreRoundTrip(t, g, true)
			me = g.Seats[g.Turn.ActiveSeat]
		}
		kharnAnswer(t, g, me.ID, third.ID)
		if me.Hand.Size() != hand+1 {
			t.Fatalf("restored=%v: hand %d after the answer, want %d", restored, me.Hand.Size(), hand+1)
		}
		if got := kharnController(t, g, kharn); got != third.ID {
			t.Fatalf("restored=%v: Khârn's controller %s, want %s", restored, got, third.ID)
		}
	}
}

// Vivi Ornitier — "put a +1/+1 counter on Vivi Ornitier and it deals 1
// damage to each opponent" — with Doubling Season and Hardened Scales:
// the counter waits on the CR 616 order, and so does the damage. An
// undo taken while the order is open keeps the damage parked too. (The
// CR 616 resume itself is a closure, so this table is not a restore
// point; the parked steps add nothing to the census.)
func TestViviDamageWaitsForTheCounterOrder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hs := seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me.ID)
	ds := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	vivi := b12Push(g, me.ID, "Vivi Ornitier", "Legendary Creature — Wizard", viviOrnitierOracle, 0, 3)
	life := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		life[p.ID] = p.Life
	}
	castCatalogSpell(t, g, "Sandbox Instant", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("want the one CR 616 prompt open, got %d choices", len(g.PendingChoices))
	}
	for _, p := range g.Seats {
		if p.Life != life[p.ID] {
			t.Fatalf("seat %s life %d with the order open, want %d: the damage ran before the counter", p.Name, p.Life, life[p.ID])
		}
	}
	snap := g.CaptureSnapshot()
	if n := snap.Continuations.ChoiceResumeFrames; n != 1 {
		t.Fatalf("census counts %d resume frames, want 1 (the CR 616 resume alone): %+v", n, snap.Continuations)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		answerOrderBySource(t, table, ds, hs)
		if got := counterCount(table, vivi, game.CounterPlusOne); got != 3 {
			t.Fatalf("+1/+1 counters %d, want 3 ((1×2)+1, Doubling Season first)", got)
		}
		for _, p := range table.Seats {
			want := life[p.ID]
			if p.ID != me.ID {
				want--
			}
			if p.Life != want {
				t.Fatalf("seat %s life %d after the order, want %d", p.Name, p.Life, want)
			}
		}
	}
}

// "Scry 2, then draw a card" written as two steps with no Then: the draw
// takes the card the scry left on top, also on a table restored while
// the scry is open.
func TestDoDrawAfterScryWaitsForTheScry(t *testing.T) {
	for _, restored := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		ids := seedLibrary(me, "Top", "Second")
		hand := me.Hand.Size()
		doWalkProbe(t, g, me.ID, 1, uuid.Nil)
		if scryChoiceFor(g, me.ID) == nil {
			t.Fatal("the scry is not open")
		}
		if me.Hand.Size() != hand {
			t.Fatalf("hand %d with the scry open, want %d: the draw ran before the scry", me.Hand.Size(), hand)
		}
		if restored {
			g = restoreRoundTrip(t, g, true)
			me = g.Seats[g.Turn.ActiveSeat]
		}
		c := scryChoiceFor(g, me.ID)
		if err := g.ResolveScry(c.ID, me.ID, []uuid.UUID{ids[0]}, []uuid.UUID{ids[1]}); err != nil {
			t.Fatal(err)
		}
		if me.Hand.Size() != hand+1 || !me.Hand.Contains(ids[1]) {
			t.Fatalf("restored=%v: hand %d, holds Second %v; want %d and true — the draw takes the card the scry kept on top",
				restored, me.Hand.Size(), me.Hand.Contains(ids[1]), hand+1)
		}
	}
}

// A step inside an If's steps parks one level down: the draw inside the
// If and the life gain after it both wait for the scry, across a
// restore, and the condition is not asked again.
func TestDoStepsInsideAnIfWaitForTheScry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedLibrary(me, "Top", "Second")
	hand, life := me.Hand.Size(), me.Life
	doWalkProbe(t, g, me.ID, 2, uuid.Nil)
	if me.Hand.Size() != hand || me.Life != life {
		t.Fatalf("hand %d, life %d with the scry open; want %d, %d", me.Hand.Size(), me.Life, hand, life)
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("restored: the scry is not open")
	}
	if err := g.ResolveScry(c.ID, me.ID, []uuid.UUID{ids[0]}, nil); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(ids[1]) || me.Hand.Size() != hand+1 || me.Life != life+1 {
		t.Fatalf("hand %d (Second %v), life %d; want %d (true), %d", me.Hand.Size(), me.Hand.Contains(ids[1]), me.Life, hand+1, life+1)
	}
}
