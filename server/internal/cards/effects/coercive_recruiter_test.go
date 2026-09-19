package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const coerciveRecruiterOracle = "ad33530c-a8df-4c1c-a863-501e583290b6"

// pushTappedVanillaCreature is pushVanillaCreature with the permanent
// already tapped, so a test can pin the untap clause.
func pushTappedVanillaCreature(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	id := pushVanillaCreature(g, owner, name, power, toughness)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
	return id
}

// TestCoerciveRecruiterStealsUntapsHastesAndAddsPirate exercises every
// printed clause: another Pirate entering under your control triggers
// it, and the resolution steals a tapped opposing creature, untaps
// it, grants haste and adds the Pirate type — all until end of turn.
func TestCoerciveRecruiterStealsUntapsHastesAndAddsPirate(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	recruiter := pushDiesCreatureForTest(g, me.ID, "Coercive Recruiter", coerciveRecruiterOracle,
		"Creature — Orc Pirate", 4, 3)
	bear := pushTappedVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	// Casting another Pirate you control is the trigger condition.
	castWithCost(t, g, "Test Pirate", "Creature — Human Pirate", "{1}", "")
	passPriorityAroundTable(t, g)

	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	c, ok := battlefieldCard(g, bear)
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
	assertKeywords(t, g, bear, "haste")
	if got := effectiveSubtypes(t, g, bear); !containsString(got, "Pirate") {
		t.Errorf("the stolen creature must become a Pirate in addition to its other types: %v", got)
	}
	_ = recruiter
}
