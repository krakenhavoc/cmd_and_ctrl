package testscm

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

const coerciveRecruiterOracle = "ad33530c-a8df-4c1c-a863-501e583290b6"

// TestCoerciveRecruiterStealsUntapsHastesAndAddsPirate exercises every
// printed clause: another Pirate entering under your control triggers
// it, and the resolution steals a tapped opposing creature, untaps
// it, grants haste and adds the Pirate type — all until end of turn.
func TestCoerciveRecruiterStealsUntapsHastesAndAddsPirate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	recruiter := effectstest.PushDiesCreatureForTest(g, me.ID, "Coercive Recruiter", coerciveRecruiterOracle,
		"Creature — Orc Pirate", 4, 3)
	bear := effectstest.PushTappedVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	// Casting another Pirate you control is the trigger condition.
	effectstest.CastWithCost(t, g, "Test Pirate", "Creature — Human Pirate", "{1}", "")
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	c, ok := effectstest.BattlefieldCard(g, bear)
	if !ok {
		t.Fatalf("the stolen creature must still be on the battlefield")
	}
	if c.Controller != me.ID {
		t.Errorf("control did not change: controller %s, want %s", c.Controller, me.ID)
	}
	if c.Owner != opp.ID {
		t.Errorf("ownership must not change: owner %s, want %s", c.Owner, opp.ID)
	}
	if c.Tapped {
		t.Errorf("the stolen creature must be untapped")
	}
	effectstest.AssertKeywords(t, g, bear, "haste")
	if got := effectstest.EffectiveSubtypes(t, g, bear); !effectstest.ContainsString(got, "Pirate") {
		t.Errorf("the stolen creature must become a Pirate in addition to its other types: %v", got)
	}
	_ = recruiter
}
