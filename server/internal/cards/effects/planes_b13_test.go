package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_b13_test.go — Planechase plane cards, batches 13–15 (ADR 0216
// slice 4, #4870): the gold-bordered Unknown Event planes. Every test plays
// the real catalog entry face up on a Planechase table (planes_harness_test.go).

var planeTarnation = planeFixture{"08081904-4115-4afd-ac82-f00421e54038", "Tarnation", "Plane — Thunder Junction"}

func TestTarnationLetsTheCriminalDrawAndChaosPings(t *testing.T) {
	requirePlaneRegisteredFull(t, planeTarnation)
	g, plane := newPlaneGame(t, planeTarnation)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	foe := opponentOf(g)
	before := me.Hand.Size()
	// Targeting an opponent is a crime (CR 700.13).
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: foe.ID}})
	if n := itemsFromPlane(g, plane.InstanceID); n != 1 {
		t.Fatalf("the crime put %d plane items on the way, want 1", n)
	}
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	// The bolt left the hand, one card was drawn.
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d after casting a bolt (a card put in it) and drawing one, want %d", got, before+1)
	}

	life := foe.Life
	rollBlank(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatalf("a blank triggered the plane")
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPlayerTarget(t, g, foe.ID)
	// Targeting an opponent is itself a crime: Tarnation asks the planar
	// controller again, who draws on the way.
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if foe.Life != life-1 {
		t.Errorf("foe life %d, want %d after 1 damage", foe.Life, life-1)
	}

	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	rollChaos(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Error("a face-down plane's chaos ability triggered")
	}
}

func answerPlayerTarget(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: who}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

var planeSkyDeck = planeFixture{"8baa42cf-32bc-4d1f-b4bd-02a4fce376c4", "Sky Deck", "Plane — Chicago"}

func TestSkyDeckPutsGlassCountersAndDestroysThemOnTheWayOut(t *testing.T) {
	requirePlaneRegisteredFull(t, planeSkyDeck)
	g, plane := newPlaneGame(t, planeSkyDeck)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	foe := opponentOf(g)
	mine := pushBattlefieldCreature(g, me.ID)
	theirs := pushBattlefieldCreature(g, foe.ID)
	bystander := pushBattlefieldCreature(g, foe.ID)
	g.BumpLayerVersionForTest()

	rollBlank(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatalf("a blank triggered the plane")
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, theirs); !ok || c.Counters[planesB13GlassCounter] != 1 {
		t.Fatalf("the chosen permanent has %v glass counters, want 1", c.Counters)
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, mine)
	passPriorityAroundTable(t, g)

	// Walking away destroys every permanent with a glass counter.
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, theirs); ok {
		t.Error("the opponent's glass creature survived the planeswalk away")
	}
	if _, ok := battlefieldCard(g, mine); ok {
		t.Error("the planar controller's glass creature survived")
	}
	if _, ok := battlefieldCard(g, bystander); !ok {
		t.Error("a creature with no glass counter was destroyed")
	}
}
