package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func arceeRow() cards.Card {
	return transformerRow(arceeOracleID,
		disturbFace{"Arcee, Sharpshooter", "Legendary Artifact Creature — Robot", "{1}{R}{W}", "2", "2"},
		disturbFace{"Arcee, Acrobatic Coupe", "Legendary Artifact — Vehicle", "", "2", "2"},
		"More Than Meets the Eye {R}{W}\nFirst strike\n{1}, Remove one or more +1/+1 counters from Arcee: It deals that much damage to target creature. Convert Arcee.",
		"Living metal\nWhenever you cast a spell that targets one or more creatures or Vehicles you control, put that many +1/+1 counters on Arcee. Convert Arcee.",
		"Living metal", "Convert", "First strike")
}

// Remove N counters, deal N damage, convert.
func TestArceeRemovesCountersToDealThatMuchDamageAndConverts(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerIn(t, g, arceeRow(), me)
	pfAddCounters(g, id, "+1/+1", 3)
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 3)
	b06AddMana(me, "C")

	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{
		CounterCounts: []int{2}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})

	c := wantFace(t, g, id, 1, "Arcee, Acrobatic Coupe")
	if c.Counters["+1/+1"] != 1 {
		t.Errorf("%d counters left, want 1 (3 minus the 2 removed)", c.Counters["+1/+1"])
	}
	b, _ := battlefieldCard(g, bear)
	if b.DamageMarked != 2 {
		t.Errorf("the bear took %d, want exactly the 2 counters removed", b.DamageMarked)
	}
}

// "One or more": removing none is refused.
func TestArceeNeedsAtLeastOneCounter(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerIn(t, g, arceeRow(), me)
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 3)
	b06AddMana(me, "C")
	err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		CounterCounts: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if err == nil {
		t.Fatal("activated with no counters to remove")
	}
	wantFace(t, g, id, 0, "Arcee, Sharpshooter")
}

// The Vehicle puts a counter for each creature or Vehicle you control the
// spell targets, then converts.
func TestArceeAcrobaticCoupeCountsTargetsYouControl(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerConverted(t, g, arceeRow(), me)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	vehicle := pushCatalogPermanent(g, me.ID, "My Vehicle", "Artifact — Vehicle", "", false)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	castCatalogSpell(t, g, "Double Shot", "Instant", "", []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: vehicle}, {Kind: game.TargetCard, ID: theirs},
	})
	tfSettle(t, g)

	c := wantFace(t, g, id, 0, "Arcee, Sharpshooter")
	if c.Counters["+1/+1"] != 2 {
		t.Errorf("%d counters, want 2 (my creature and my Vehicle, not theirs)", c.Counters["+1/+1"])
	}
}

// A spell that targets nothing of yours does not trigger it, and neither
// does an opponent's spell.
func TestArceeAcrobaticCoupeIgnoresOtherSpells(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerConverted(t, g, arceeRow(), me)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	castCatalogSpell(t, g, "Shot", "Instant", "", []game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	tfSettle(t, g)
	wantFace(t, g, id, 1, "Arcee, Acrobatic Coupe")
	_ = uuid.Nil
}
