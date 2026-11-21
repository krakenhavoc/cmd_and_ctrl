package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardsap "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ap"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func arceeRow() cards.Card {
	return effectstest.TransformerRow(cardsap.ArceeOracleID,
		effectstest.DisturbFace{Name: "Arcee, Sharpshooter", TypeLine: "Legendary Artifact Creature — Robot", Cost: "{1}{R}{W}", Power: "2", Toughness: "2"},
		effectstest.DisturbFace{Name: "Arcee, Acrobatic Coupe", TypeLine: "Legendary Artifact — Vehicle", Cost: "", Power: "2", Toughness: "2"},
		"More Than Meets the Eye {R}{W}\nFirst strike\n{1}, Remove one or more +1/+1 counters from Arcee: It deals that much damage to target creature. Convert Arcee.",
		"Living metal\nWhenever you cast a spell that targets one or more creatures or Vehicles you control, put that many +1/+1 counters on Arcee. Convert Arcee.",
		"Living metal", "Convert", "First strike")
}

// Remove N counters, deal N damage, convert.
func TestArceeRemovesCountersToDealThatMuchDamageAndConverts(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, arceeRow(), me)
	effectstest.PfAddCounters(g, id, "+1/+1", 3)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 3)
	effectstest.B06AddMana(me, "C")

	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{
		CounterCounts: []int{2}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})

	c := effectstest.WantFace(t, g, id, 1, "Arcee, Acrobatic Coupe")
	if c.Counters["+1/+1"] != 1 {
		t.Errorf("%d counters left, want 1 (3 minus the 2 removed)", c.Counters["+1/+1"])
	}
	b, _ := effectstest.BattlefieldCard(g, bear)
	if b.DamageMarked != 2 {
		t.Errorf("the bear took %d, want exactly the 2 counters removed", b.DamageMarked)
	}
}

// "One or more": removing none is refused.
func TestArceeNeedsAtLeastOneCounter(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, arceeRow(), me)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 3)
	effectstest.B06AddMana(me, "C")
	err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		CounterCounts: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if err == nil {
		t.Fatal("activated with no counters to remove")
	}
	effectstest.WantFace(t, g, id, 0, "Arcee, Sharpshooter")
}

// The Vehicle puts a counter for each creature or Vehicle you control the
// spell targets, then converts.
func TestArceeAcrobaticCoupeCountsTargetsYouControl(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.TransformerConverted(t, g, arceeRow(), me)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	vehicle := effectstest.PushCatalogPermanent(g, me.ID, "My Vehicle", "Artifact — Vehicle", "", false)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Double Shot", "Instant", "", []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: vehicle}, {Kind: game.TargetCard, ID: theirs},
	})
	effectstest.TfSettle(t, g)

	c := effectstest.WantFace(t, g, id, 0, "Arcee, Sharpshooter")
	if c.Counters["+1/+1"] != 2 {
		t.Errorf("%d counters, want 2 (my creature and my Vehicle, not theirs)", c.Counters["+1/+1"])
	}
}

// A spell that targets nothing of yours does not trigger it, and neither
// does an opponent's spell.
func TestArceeAcrobaticCoupeIgnoresOtherSpells(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.TransformerConverted(t, g, arceeRow(), me)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Shot", "Instant", "", []game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.TfSettle(t, g)
	effectstest.WantFace(t, g, id, 1, "Arcee, Acrobatic Coupe")
	_ = uuid.Nil
}
