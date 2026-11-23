package testscm

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

const conjurersClosetOracle = "cd1eda60-53e4-44d0-9b2c-7a57395e291f"

// TestConjurersClosetBlinksACreatureYouControlAtYourEndStep — same
// shape as Thassa, Deep-Dwelling's end-step blink: the target is picked
// as the trigger goes on the stack, the "you may" is answered as it
// resolves (CR 603.5), and the exiled creature
// comes back as a new object under the controller's control.
func TestConjurersClosetBlinksACreatureYouControlAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Conjurer's Closet", "Artifact", conjurersClosetOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	effectstest.AdvanceToEndStepOf(t, g, 0)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick_target prompt at your end step: %+v", g.PendingChoices)
	}
	if !effectstest.HasID(pick.PickTargetCards, bear) {
		t.Fatal("a creature you control is not offered")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	yn := effectstest.AwaitTriggerPrompt(t, g, me.ID)
	if err := g.ResolveTriggerPrompt(yn.ID, me.ID, true); err != nil {
		t.Fatalf("answering yes: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	back := effectstest.FindBattlefieldByName(g, "Bear")
	if back == uuid.Nil || back == bear {
		t.Fatalf("blinked creature back=%v (old %v): want a new object on the battlefield", back, bear)
	}
	if c, _ := effectstest.BattlefieldCard(g, back); c.Controller != me.ID {
		t.Errorf("returned under %v's control, want yours (%v)", c.Controller, me.ID)
	}
}

// "You may": a "No" answer leaves the creature alone.
func TestConjurersClosetMayDeclineTheBlink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Conjurer's Closet", "Artifact", conjurersClosetOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PickCard(t, g, me.ID, bear)
	yn := effectstest.AwaitTriggerPrompt(t, g, me.ID)
	if err := g.ResolveTriggerPrompt(yn.ID, me.ID, false); err != nil {
		t.Fatalf("declining: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !effects.OnBattlefield(g, bear) {
		t.Error("declining still moved the creature")
	}
}
