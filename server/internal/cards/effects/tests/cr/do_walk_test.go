package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func doWalkProbe(t *testing.T, g *game.Game, me uuid.UUID, idx int, target uuid.UUID) {
	t.Helper()
	probe := effectstest.PushCatalogPermanent(g, me, "Do Walk Probe", "Artifact", effectstest.DoWalkProbeOracle, false)
	if err := effectstest.ActivateOn(t, g, me, probe, idx, target); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// Khârn the Betrayer is dealt damage by the first step of a Do; the
// prevention follow-up asks its controller for an opponent (#2040).
// "Draw a card" is the next instruction, so it waits for the answer —
// on the live table and on one restored while the question is open.
func TestDoStepAfterKharnsQuestionWaitsForTheAnswer(t *testing.T) {
	for _, restored := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, third := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
		kharn := effectstest.KharnPush(g, me.ID, me.ID)
		hand := me.Hand.Size()
		doWalkProbe(t, g, me.ID, 0, kharn)
		if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil {
			t.Fatal("Khârn's question is not open")
		}
		if me.Hand.Size() != hand {
			t.Fatalf("hand %d with the question open, want %d: the draw ran before the answer", me.Hand.Size(), hand)
		}
		if restored {
			g = effectstest.RestoreRoundTrip(t, g, true)
			me = g.Seats[g.Turn.ActiveSeat]
		}
		effectstest.KharnAnswer(t, g, me.ID, third.ID)
		if me.Hand.Size() != hand+1 {
			t.Fatalf("restored=%v: hand %d after the answer, want %d", restored, me.Hand.Size(), hand+1)
		}
		if got := effectstest.KharnController(t, g, kharn); got != third.ID {
			t.Fatalf("restored=%v: Khârn's controller %s, want %s", restored, got, third.ID)
		}
	}
}

// Vivi Ornitier — "put a +1/+1 counter on Vivi Ornitier and it deals 1
// damage to each opponent" — with Doubling Season and Hardened Scales:
// the counter waits on the CR 616 order, and so does the damage. An
// undo taken while the order is open keeps the damage parked too, and
// so does a strict restore: the counter placement has no tail, so its
// CR 616 frame is data (#4253), and the parked steps add nothing to the
// census.
func TestViviDamageWaitsForTheCounterOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hs := effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me.ID)
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	vivi := effectstest.B12Push(g, me.ID, "Vivi Ornitier", "Legendary Creature — Wizard", effectstest.ViviOrnitierOracle, 0, 3)
	life := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		life[p.ID] = p.Life
	}
	effectstest.CastCatalogSpell(t, g, "Sandbox Instant", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("want the one CR 616 prompt open, got %d choices", len(g.PendingChoices))
	}
	for _, p := range g.Seats {
		if p.Life != life[p.ID] {
			t.Fatalf("seat %s life %d with the order open, want %d: the damage ran before the counter", p.Name, p.Life, life[p.ID])
		}
	}
	undo := g.Clone()
	restored := effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCounter)
	for _, table := range []*game.Game{g, undo, restored} {
		effectstest.AnswerOrderBySource(t, table, ds, hs)
		if got := effectstest.CounterCount(table, vivi, game.CounterPlusOne); got != 3 {
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		ids := effectstest.SeedLibrary(me, "Top", "Second")
		hand := me.Hand.Size()
		doWalkProbe(t, g, me.ID, 1, uuid.Nil)
		if effectstest.ScryChoiceFor(g, me.ID) == nil {
			t.Fatal("the scry is not open")
		}
		if me.Hand.Size() != hand {
			t.Fatalf("hand %d with the scry open, want %d: the draw ran before the scry", me.Hand.Size(), hand)
		}
		if restored {
			g = effectstest.RestoreRoundTrip(t, g, true)
			me = g.Seats[g.Turn.ActiveSeat]
		}
		c := effectstest.ScryChoiceFor(g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedLibrary(me, "Top", "Second")
	hand, life := me.Hand.Size(), me.Life
	doWalkProbe(t, g, me.ID, 2, uuid.Nil)
	if me.Hand.Size() != hand || me.Life != life {
		t.Fatalf("hand %d, life %d with the scry open; want %d, %d", me.Hand.Size(), me.Life, hand, life)
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	c := effectstest.ScryChoiceFor(g, me.ID)
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

// Khârn the Betrayer and Selfless Squire, dealt damage by one instance
// in a Do's first step: Khârn asks, and the Squire's announcement (#2047,
// CR 615.13) waits behind the question (#2040). The next step waits too,
// and when it runs the announcement has already been emitted, because
// it belongs to the earlier step's damage (CR 615.5, CR 608.2c). Live
// and across a strict restore taken while the question is open.
func TestDoStepAfterKharnWaitsForTheSquiresAnnouncement(t *testing.T) {
	for _, restored := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, third := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
		kharn := effectstest.KharnPush(g, me.ID, me.ID)
		effectstest.DoWalkSquire = effectstest.CastSquire(t, g)
		effectstest.DoWalkSquireSeen = -1
		life := me.Life
		doWalkProbe(t, g, me.ID, 3, kharn)
		if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil {
			t.Fatal("Khârn's question is not open")
		}
		if n := effectstest.SquireTriggers(g, effectstest.DoWalkSquire); n != 0 || effectstest.DoWalkSquireSeen != -1 {
			t.Fatalf("%d Squire triggers and the next step saw %d with the question open; want 0 and not run", n, effectstest.DoWalkSquireSeen)
		}
		if restored {
			g = effectstest.RestoreRoundTrip(t, g, true)
			me = g.Seats[g.Turn.ActiveSeat]
		}
		effectstest.KharnAnswer(t, g, me.ID, third.ID)
		if effectstest.DoWalkSquireSeen != 1 {
			t.Fatalf("restored=%v: the next step saw %d Squire triggers, want 1 — the announcement first", restored, effectstest.DoWalkSquireSeen)
		}
		if n := effectstest.SquireTriggers(g, effectstest.DoWalkSquire); n != 1 {
			t.Fatalf("restored=%v: %d Squire triggers after the answer, want 1", restored, n)
		}
		effectstest.SettleOrderingTriggers(t, g)
		if n := effectstest.CountersOn(g, effectstest.DoWalkSquire, game.CounterPlusOne); n != 2 || me.Life != life {
			t.Fatalf("restored=%v: Squire counters %d (want 2), life %d (want %d)", restored, n, me.Life, life)
		}
		if got := effectstest.KharnController(t, g, kharn); got != third.ID {
			t.Fatalf("restored=%v: Khârn's controller %s, want %s", restored, got, third.ID)
		}
	}
}
