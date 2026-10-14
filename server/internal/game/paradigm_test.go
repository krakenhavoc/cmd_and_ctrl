package game

import (
	"testing"

	"github.com/google/uuid"
)

// paradigm_test.go — paradigm (CR 702.192), #1866.
//
// What a bug here would hide, worst first:
//
//  1. The record. Only the FIRST resolution of a name schedules the
//     trigger; a second card of the same name, or the copy the trigger
//     offers, must not add another free spell every turn.
//  2. The gate. A countered or fizzled paradigm spell did not resolve,
//     so it goes to the graveyard and creates nothing.
//  3. The trigger's turn. "Each of YOUR precombat main phases", and
//     every one of them for the rest of the game.
//  4. The restore point. The trigger is data.

// seedParadigmCard puts a paradigm sorcery in the seat's hand. No
// catalog entry: the keyword is the card's own, as a deck import
// stamps it.
func seedParadigmCard(p *Player, name string) Card {
	c := NewCard(name, p.ID)
	c.TypeLine = "Sorcery — Lesson"
	c.ManaCost = "{3}{B}{B}"
	c.Keywords = []string{KeywordParadigm}
	p.Hand.PushTop(c)
	return c
}

// castAndResolveParadigm casts a paradigm card from hand in the active
// seat's main phase and resolves it.
func castAndResolveParadigm(t *testing.T, g *Game, me *Player, name string) uuid.UUID {
	t.Helper()
	advanceTo(t, g, StepPrecombatMain)
	c := seedParadigmCard(me, name)
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast from hand: %v", err)
	}
	resolveTop(t, g)
	return c.InstanceID
}

// paradigmTriggers is the queued paradigm delayed triggers.
func paradigmTriggers(g *Game) []*DelayedTrigger {
	var out []*DelayedTrigger
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.Body.Key() == "paradigm/copy" {
			out = append(out, dt)
		}
	}
	return out
}

// paradigmCopiesInExile is the card copies sitting in exile.
func paradigmCopiesInExile(g *Game) []Card {
	var out []Card
	for _, c := range g.Exile.Cards {
		if c.CardCopy {
			out = append(out, c)
		}
	}
	return out
}

// reachParadigmOffer walks to the seat's next precombat main and passes
// until the trigger has resolved and a copy is on offer.
func reachParadigmOffer(t *testing.T, g *Game, seat int) Card {
	t.Helper()
	for g.Turn.ActiveSeat == seat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	advanceToStepOfSeat(t, g, seat, StepPrecombatMain)
	passUntil(t, g, func() bool { return len(paradigmCopiesInExile(g)) > 0 })
	return paradigmCopiesInExile(g)[0]
}

// CR 702.192a: the spell exiles itself, and its first resolution
// creates one delayed trigger for its controller's precombat main
// phases, for the rest of the game.
func TestParadigmSpellIsExiledAndSchedulesTheTrigger(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := castAndResolveParadigm(t, g, me, "Test Paradigm")

	if me.Graveyard.Contains(id) || !g.Exile.Contains(id) {
		t.Fatal("a resolved paradigm spell was not exiled")
	}
	dts := paradigmTriggers(g)
	if len(dts) != 1 {
		t.Fatalf("paradigm triggers: got %d, want 1", len(dts))
	}
	dt := dts[0]
	if dt.At != StepPrecombatMain || !dt.ControllerTurnOnly || dt.Controller != me.ID {
		t.Errorf("trigger = at %q, controller-turn-only %v, controller %v; want precombat main, true, %v",
			dt.At, dt.ControllerTurnOnly, dt.Controller, me.ID)
	}
	if dt.Duration == nil || dt.Duration.Kind != Indefinite {
		t.Errorf("trigger duration %+v, want the rest of the game", dt.Duration)
	}
	if dt.Params.Object.ID != id || dt.Params.Name != "Test Paradigm" {
		t.Errorf("trigger params %+v, want the exiled card and its name", dt.Params)
	}
}

// "If this is the first time a spell you control with this spell's
// name has resolved this game": a second card of the same name is
// exiled and creates nothing; another name starts its own record.
func TestParadigmRecordIsPerName(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	castAndResolveParadigm(t, g, me, "Test Paradigm")
	second := castAndResolveParadigm(t, g, me, "Test Paradigm")
	if !g.Exile.Contains(second) {
		t.Error("the second card of the name was not exiled")
	}
	if n := len(paradigmTriggers(g)); n != 1 {
		t.Fatalf("paradigm triggers after two of one name: %d, want 1", n)
	}
	castAndResolveParadigm(t, g, me, "Other Paradigm")
	if n := len(paradigmTriggers(g)); n != 2 {
		t.Fatalf("paradigm triggers after a second name: %d, want 2", n)
	}
}

// CR 608.2n: a countered paradigm spell never resolved. It goes to the
// graveyard and creates nothing.
func TestCounteredParadigmSpellGoesToTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	c := seedParadigmCard(me, "Test Paradigm")
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if err := g.CounterSpell(c.InstanceID, nil); err != nil {
		t.Fatalf("CounterSpell: %v", err)
	}
	if !me.Graveyard.Contains(c.InstanceID) || g.Exile.Contains(c.InstanceID) {
		t.Error("a countered paradigm spell was not put into its owner's graveyard")
	}
	if n := len(paradigmTriggers(g)); n != 0 {
		t.Errorf("a countered paradigm spell created %d delayed triggers", n)
	}
}

// The trigger waits through an opponent's precombat main, fires on the
// controller's, offers a free copy, and stays queued for the next one.
// The copy is cast for {0}, resolves, and ceases to exist; it neither
// exiles the original nor starts a second record.
func TestParadigmOffersAFreeCopyEachOfYourPrecombatMains(t *testing.T) {
	g := newActiveGame(t)
	if len(g.Seats) < 2 {
		t.Skip("needs two seats")
	}
	me := g.Seats[0]
	id := castAndResolveParadigm(t, g, me, "Test Paradigm")

	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	if len(g.PendingTriggers)+len(g.StackMeta) != 0 || len(paradigmCopiesInExile(g)) != 0 {
		t.Fatal("the trigger fired on an opponent's precombat main")
	}

	for turn := 0; turn < 2; turn++ {
		cp := reachParadigmOffer(t, g, 0)
		if cp.Name != "Test Paradigm" || cp.InstanceID == id {
			t.Fatalf("turn %d: copy %+v, want a new object named Test Paradigm", turn, cp)
		}
		if err := g.CastSpell(me.ID, cp.InstanceID, CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
			t.Fatalf("turn %d: free cast of the copy: %v", turn, err)
		}
		if item := g.StackMeta[cp.InstanceID]; item == nil || !item.IsCopy {
			t.Fatalf("turn %d: the cast copy is not a copy spell on the stack", turn)
		}
		resolveTop(t, g)
		if n := len(paradigmCopiesInExile(g)); n != 0 {
			t.Errorf("turn %d: %d copies left in exile after the copy resolved", turn, n)
		}
		if !g.Exile.Contains(id) {
			t.Errorf("turn %d: the original left exile", turn)
		}
		if n := len(paradigmTriggers(g)); n != 1 {
			t.Fatalf("turn %d: paradigm triggers %d, want the one record", turn, n)
		}
	}
}

// Passing with the copy uncast declines it: the copy ceases to exist
// and the original stays in exile for the next turn's offer.
func TestParadigmCopyNotCastCeasesToExist(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := castAndResolveParadigm(t, g, me, "Test Paradigm")
	reachParadigmOffer(t, g, 0)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if n := len(paradigmCopiesInExile(g)); n != 0 {
		t.Errorf("%d copies left after the holder passed", n)
	}
	if !g.Exile.Contains(id) || len(paradigmTriggers(g)) != 1 {
		t.Error("declining the copy cost the original or the record")
	}
}

// The trigger is data: a table holding one survives a snapshot, and the
// restored game makes the offer.
func TestParadigmSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := castAndResolveParadigm(t, g, me, "Test Paradigm")

	_, restored := roundTrip(t, g)
	if !restored.Exile.Contains(id) {
		t.Fatal("the exiled paradigm card did not survive the round trip")
	}
	if n := len(paradigmTriggers(restored)); n != 1 {
		t.Fatalf("paradigm triggers after the round trip: %d, want 1", n)
	}
	if cp := reachParadigmOffer(t, restored, 0); cp.Name != "Test Paradigm" {
		t.Errorf("restored offer is a copy of %q", cp.Name)
	}
}
