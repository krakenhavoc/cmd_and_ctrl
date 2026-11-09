package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_b9_test.go — Planechase plane cards, batch 9 (Planechase Anthology
// N–S, ADR 0216 slice 4, #4870). Each test plays the real catalog entry
// face up on a Planechase table (planes_harness_test.go).

var (
	planeNephalia = planeFixture{"72af7529-75bb-488f-801d-7148631b9e9b", "Nephalia", "Plane — Innistrad"}
	planeTazeem   = planeFixture{"97f708de-9271-4137-a9fe-dda25b922900", "Tazeem", "Plane — Zendikar"}
)

// planeLand puts a basic land on the battlefield under `owner`.
func planeLand(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: owner, Controller: owner,
	})
}

// --- Nephalia ----------------------------------------------------------------

func TestNephaliaMillsSevenThenReturnsOneAtRandomAtYourEnd(t *testing.T) {
	requirePlaneRegisteredFull(t, planeNephalia)
	g, plane := newPlaneGame(t, planeNephalia)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	lib, hand, yard := me.Library.Size(), me.Hand.Size(), me.Graveyard.Size()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != lib-7 {
		t.Errorf("library %d, want %d after milling seven", got, lib-7)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d: one milled card returned", got, hand+1)
	}
	if got := me.Graveyard.Size(); got != yard+6 {
		t.Errorf("graveyard %d, want %d", got, yard+6)
	}
	_ = plane
}

func TestNephaliaChaosReturnsTargetCard(t *testing.T) {
	g, plane := newPlaneGame(t, planeNephalia)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushGraveyardCardWithTypeLineID(me, "Bear", "Creature — Bear")
	rollBlank(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatal("a blank triggered the chaos ability")
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the targeted card did not return to hand")
	}
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	rollChaos(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Error("a face-down plane's chaos ability triggered")
	}
}

// --- Tazeem ------------------------------------------------------------------

func TestTazeemStopsBlocksAndChaosDrawsPerLand(t *testing.T) {
	requirePlaneRegisteredFull(t, planeTazeem)
	g, _ := newPlaneGame(t, planeTazeem)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	for range 3 {
		planeLand(g, me.ID)
	}
	planeLand(g, opp.ID)
	raider := b12Creature(g, me.ID, "Raider", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	before := me.Hand.Size()
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+3 {
		t.Errorf("hand %d, want %d: a card per land you control", got, before+3)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(raider, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, raider); err == nil {
		t.Error("a creature blocked on Tazeem")
	}
}
