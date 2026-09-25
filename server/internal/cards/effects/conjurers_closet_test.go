package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const conjurersClosetOracle = "cd1eda60-53e4-44d0-9b2c-7a57395e291f"

// TestConjurersClosetBlinksACreatureYouControlAtYourEndStep — same
// shape as Thassa, Deep-Dwelling's end-step blink: the "you may" is
// answered first, then the target is picked, and the exiled creature
// comes back as a new object under the controller's control.
func TestConjurersClosetBlinksACreatureYouControlAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Conjurer's Closet", "Artifact", conjurersClosetOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	advanceToEndStepOf(t, g, 0)
	yn := latestChoiceOfKindFor(g, game.PendingChoiceTriggerPrompt, me.ID)
	if yn == nil {
		t.Fatalf("no \"you may\" prompt at your end step: %+v", g.PendingChoices)
	}
	if err := g.ResolveTriggerPrompt(yn.ID, me.ID, true); err != nil {
		t.Fatalf("answering yes: %v", err)
	}

	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick_target prompt after answering yes: %+v", g.PendingChoices)
	}
	if !hasID(pick.PickTargetCards, bear) {
		t.Fatal("a creature you control is not offered")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	back := findBattlefieldByName(g, "Bear")
	if back == uuid.Nil || back == bear {
		t.Fatalf("blinked creature back=%v (old %v): want a new object on the battlefield", back, bear)
	}
	if c, _ := battlefieldCard(g, back); c.Controller != me.ID {
		t.Errorf("returned under %v's control, want yours (%v)", c.Controller, me.ID)
	}
}

// "You may": a "No" answer leaves the creature alone.
func TestConjurersClosetMayDeclineTheBlink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Conjurer's Closet", "Artifact", conjurersClosetOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	advanceToEndStepOf(t, g, 0)
	yn := latestChoiceOfKindFor(g, game.PendingChoiceTriggerPrompt, me.ID)
	if yn == nil {
		t.Fatalf("no \"you may\" prompt at your end step: %+v", g.PendingChoices)
	}
	if err := g.ResolveTriggerPrompt(yn.ID, me.ID, false); err != nil {
		t.Fatalf("declining: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !onBattlefield(g, bear) {
		t.Error("declining still moved the creature")
	}
}
