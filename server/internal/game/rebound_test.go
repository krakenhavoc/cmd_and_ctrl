package game

import (
	"testing"

	"github.com/google/uuid"
)

// rebound_test.go — rebound (CR 702.88), #1854, ADR 0107 §3.
//
// What a bug here would hide, worst first:
//
//  1. The gate. Rebound replaces ONE exit — "as it resolves" — and
//     only for a spell cast from a hand. A countered, fizzled or
//     exile-cast rebound card that came back anyway is a free spell
//     every turn.
//  2. The trigger's turn. "Your NEXT upkeep" is the controller's, and
//     an opponent's upkeep in between does nothing.
//  3. The object. A card that left exile before the upkeep is a new
//     object (CR 400.7) and is offered nothing.
//  4. The restore point. The exiled card and its delayed trigger are
//     data, so a table with one waiting survives a snapshot.

// seedReboundCard puts a rebound sorcery in the seat's hand. No
// catalog entry: the keyword is the card's own, as a deck import
// stamps it.
func seedReboundCard(p *Player, name, typeLine string) Card {
	c := NewCard(name, p.ID)
	c.TypeLine = typeLine
	c.ManaCost = "{1}{U}"
	c.Keywords = []string{KeywordRebound}
	p.Hand.PushTop(c)
	return c
}

// castAndResolveRebound casts a rebound card from hand at the active
// seat's main phase and resolves it.
func castAndResolveRebound(t *testing.T, g *Game, me *Player, typeLine string) uuid.UUID {
	t.Helper()
	advanceTo(t, g, StepPrecombatMain)
	c := seedReboundCard(me, "Test Rebound", typeLine)
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast from hand: %v", err)
	}
	resolveTop(t, g)
	return c.InstanceID
}

// reboundTriggers is the queued rebound delayed triggers.
func reboundTriggers(g *Game) []*DelayedTrigger {
	var out []*DelayedTrigger
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.Body.Key() == "rebound/cast" {
			out = append(out, dt)
		}
	}
	return out
}

// reboundOnStack reports whether a rebound trigger is waiting — queued
// or on the stack.
func reboundOnStack(g *Game) bool {
	for _, it := range g.PendingTriggers {
		if it != nil && it.Body == "rebound/cast" {
			return true
		}
	}
	for _, it := range g.StackMeta {
		if it != nil && it.Body == "rebound/cast" {
			return true
		}
	}
	return false
}

func mayCastOutstanding(g *Game) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceMayCast {
			return c
		}
	}
	return nil
}

// reachReboundOffer walks to the seat's next upkeep and resolves the
// rebound trigger, stopping at the may_cast prompt.
func reachReboundOffer(t *testing.T, g *Game, seat int) *PendingChoice {
	t.Helper()
	upkeepFor(t, g, seat)
	passUntil(t, g, func() bool { return mayCastOutstanding(g) != nil })
	return mayCastOutstanding(g)
}

// CR 702.88a: cast from hand and resolved, the card is exiled instead
// of going to the graveyard, and one delayed trigger is created for
// its controller's next upkeep.
func TestReboundSpellCastFromHandIsExiledOnResolution(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := castAndResolveRebound(t, g, me, "Sorcery")

	if me.Graveyard.Contains(id) {
		t.Error("a rebound spell cast from hand went to the graveyard")
	}
	if !g.Exile.Contains(id) {
		t.Fatal("a rebound spell cast from hand was not exiled")
	}
	dts := reboundTriggers(g)
	if len(dts) != 1 {
		t.Fatalf("rebound delayed triggers: got %d, want 1", len(dts))
	}
	dt := dts[0]
	if dt.At != StepUpkeep || !dt.ControllerTurnOnly || dt.Controller != me.ID {
		t.Errorf("trigger = at %q, controller-turn-only %v, controller %v; want upkeep, true, %v",
			dt.At, dt.ControllerTurnOnly, dt.Controller, me.ID)
	}
	if dt.Params.Object.ID != id || dt.Params.Object.Epoch != exiledCardByIDLocked(g, id).ObjectEpoch {
		t.Errorf("trigger names object %+v, want the exiled card", dt.Params.Object)
	}
}

// CR 702.88c: several instances are redundant — one exile, one
// trigger.
func TestTwoInstancesOfReboundAreOneTrigger(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	c := seedReboundCard(me, "Test Rebound", "Instant")
	me.Hand.Cards[len(me.Hand.Cards)-1].Keywords = []string{KeywordRebound, KeywordRebound}
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTop(t, g)
	if n := len(reboundTriggers(g)); n != 1 {
		t.Errorf("rebound delayed triggers: got %d, want 1", n)
	}
}

// CR 608.2n, 702.88a: rebound replaces only "as it resolves". A
// countered rebound spell goes to the graveyard and creates nothing.
func TestCounteredReboundSpellGoesToTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	c := seedReboundCard(me, "Test Rebound", "Instant")
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if err := g.CounterSpell(c.InstanceID, nil); err != nil {
		t.Fatalf("CounterSpell: %v", err)
	}
	if !me.Graveyard.Contains(c.InstanceID) || g.Exile.Contains(c.InstanceID) {
		t.Error("a countered rebound spell was not put into its owner's graveyard")
	}
	if n := len(reboundTriggers(g)); n != 0 {
		t.Errorf("a countered rebound spell created %d delayed triggers", n)
	}
}

// A rebound spell whose every target became illegal is countered by
// the game rules (CR 608.2b): it never resolves, so rebound never
// applies.
func TestFizzledReboundSpellGoesToTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	advanceTo(t, g, StepPrecombatMain)
	c := seedReboundCard(me, "Test Rebound", "Instant")
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if err := g.Concede(them.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	resolveTop(t, g)
	if !me.Graveyard.Contains(c.InstanceID) || g.Exile.Contains(c.InstanceID) {
		t.Error("a fizzled rebound spell was not put into its owner's graveyard")
	}
	if n := len(reboundTriggers(g)); n != 0 {
		t.Errorf("a fizzled rebound spell created %d delayed triggers", n)
	}
}

// CR 702.88a: "if this spell was cast from your hand". A rebound card
// cast from anywhere else resolves to the graveyard.
func TestReboundSpellNotCastFromHandGoesToTheGraveyard(t *testing.T) {
	const oracle = "test-rebound-graveyard"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Rebound", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{1}{U}"
	c.OracleID = oracle
	c.Keywords = []string{KeywordRebound}
	me.Graveyard.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast from graveyard: %v", err)
	}
	resolveTop(t, g)
	if !me.Graveyard.Contains(c.InstanceID) || g.Exile.Contains(c.InstanceID) {
		t.Error("a rebound spell cast from a graveyard was exiled")
	}
	if n := len(reboundTriggers(g)); n != 0 {
		t.Errorf("created %d delayed triggers, want 0", n)
	}
}

// A card without rebound is untouched: the arm reads the keyword.
func TestASpellWithoutReboundStillGoesToTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	c := seedReboundCard(me, "Plain Sorcery", "Sorcery")
	me.Hand.Cards[len(me.Hand.Cards)-1].Keywords = nil
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTop(t, g)
	if !me.Graveyard.Contains(c.InstanceID) {
		t.Error("a spell without rebound did not reach the graveyard")
	}
}

// The whole loop: the opponent's upkeep does nothing, the
// controller's next upkeep offers the cast, "yes" opens a free cast of
// a SORCERY in the upkeep (CR 608.2g), and the recast — from exile,
// not a hand — goes to the graveyard and does not rebound again.
func TestReboundOffersAFreeCastAtYourNextUpkeep(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := castAndResolveRebound(t, g, me, "Sorcery")

	// The opponent's upkeep comes first and must leave the trigger
	// queued.
	upkeepFor(t, g, 1)
	if reboundOnStack(g) || len(reboundTriggers(g)) != 1 {
		t.Fatal("the rebound trigger fired in an opponent's upkeep")
	}

	offer := reachReboundOffer(t, g, 0)
	if offer.MayCastKeyword != MayCastKeywordRebound || offer.MayCastCard != id {
		t.Fatalf("offer = keyword %q card %v, want rebound about %v", offer.MayCastKeyword, offer.MayCastCard, id)
	}
	if len(reboundTriggers(g)) != 0 {
		t.Error("the rebound trigger is still queued after it fired")
	}
	answerMayCast(t, g, me.ID, true)

	if g.Turn.Step != StepUpkeep {
		t.Fatalf("test needs the upkeep, at %q", g.Turn.Step)
	}
	// Strict: the cast must be free. The pool is empty.
	me.ManaPool = nil
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("free cast of a rebound sorcery in the upkeep: %v", err)
	}
	resolveTop(t, g)
	if !me.Graveyard.Contains(id) {
		t.Error("the rebound recast did not go to the graveyard")
	}
	if g.Exile.Contains(id) || len(reboundTriggers(g)) != 0 {
		t.Error("the rebound recast rebounded again")
	}
}

// "You MAY cast": declining leaves the card in exile with no way to
// cast it.
func TestDecliningTheReboundCastLeavesTheCardInExile(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := castAndResolveRebound(t, g, me, "Instant")
	reachReboundOffer(t, g, 0)
	answerMayCast(t, g, me.ID, false)

	if !g.Exile.Contains(id) {
		t.Fatal("declining moved the card out of exile")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "exile"}); err != ErrNoPlayPermission {
		t.Errorf("cast after declining: got %v, want ErrNoPlayPermission", err)
	}
}

// Accepting and then passing without casting is the decline, made
// late: the window closes on the pass and the card stays in exile.
func TestPassingOnTheReboundCastLeavesTheCardInExile(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := castAndResolveRebound(t, g, me, "Instant")
	reachReboundOffer(t, g, 0)
	answerMayCast(t, g, me.ID, true)
	if g.Seats[g.Turn.PriorityHolder].ID != me.ID {
		t.Fatal("test needs the rebound controller to hold priority")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if !g.Exile.Contains(id) {
		t.Fatal("passing on the rebound cast moved the card out of exile")
	}
	for _, perm := range me.CastPermissions {
		if perm.Label == ReboundFreeCastLabel {
			t.Fatal("the rebound grant outlived its holder's pass")
		}
	}
}

// CR 603.7c, 400.7: a card that left exile before the upkeep — even
// one that came straight back — is not the object the trigger is
// about, and is offered nothing.
func TestReboundTriggerIgnoresACardThatLeftExile(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := castAndResolveRebound(t, g, me, "Instant")
	if len(reboundTriggers(g)) != 1 {
		t.Fatal("setup: no rebound trigger queued")
	}
	// Out of exile and back in: a new object.
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(id); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneHand, Owner: me.ID}, ZoneRef{Kind: ZoneExile}, id); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if !g.Exile.Contains(id) {
		t.Fatal("setup: the card is not back in exile")
	}
	upkeepFor(t, g, 0)
	passUntil(t, g, func() bool { return !reboundOnStack(g) })
	if mayCastOutstanding(g) != nil {
		t.Error("a card that left exile was still offered its rebound cast")
	}
}

// The waiting trigger is data: a snapshot round trip carries the
// exiled card and the trigger, and the restored game still offers the
// cast.
func TestReboundSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := castAndResolveRebound(t, g, me, "Instant")

	_, restored := roundTrip(t, g)
	if !restored.Exile.Contains(id) {
		t.Fatal("the exiled rebound card did not survive the round trip")
	}
	if n := len(reboundTriggers(restored)); n != 1 {
		t.Fatalf("rebound triggers after the round trip: %d, want 1", n)
	}
	offer := reachReboundOffer(t, restored, 0)
	if offer.MayCastCard != id {
		t.Errorf("restored offer names %v, want %v", offer.MayCastCard, id)
	}
}
