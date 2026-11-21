package testst

import (
	"errors"
	"fmt"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const theIndomitableOracle = "276dc5c8-c8cf-4b9c-ad75-31876e6e040a"

// TestTheIndomitableDrawsForAnyCreatureYouControl — "a creature you
// control", not "this creature" or "one or more creatures": both the
// crewed Vehicle itself and an unrelated creature connecting each
// draw a card, with no batching.
func TestTheIndomitableDrawsForAnyCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	indom := effectstest.CastCatalogSpell(t, g, "The Indomitable", "Legendary Artifact — Vehicle", theIndomitableOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, indom, "trample")

	effectstest.CrewForTest(t, g, me.ID, indom, effectstest.PushCrewerForTest(g, me.ID, "Crewer", 3))
	other := effectstest.PushVanillaCreature(g, me.ID, "Other Attacker", 2, 2)

	handBefore := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, indom, opp.ID, 6)
	effectstest.DealCombatDamageToPlayer(g, other, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Fatalf("drew %d, want 2 (one per creature that connected)", got)
	}
}

// TestTheIndomitableGraveyardCastNeedsThreeTappedPiratesOrVehicles —
// the printed condition, refused when it isn't met.
func TestTheIndomitableGraveyardCastNeedsThreeTappedPiratesOrVehicles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.SeedGraveyardCard(t, g, "The Indomitable", "Legendary Artifact — Vehicle", theIndomitableOracle)
	effectstest.B06AddMana(me, "U", "U", "C", "C")

	err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard"})
	var cantCast *game.CantCastError
	if !errors.As(err, &cantCast) {
		t.Fatalf("cast with no tapped Pirates/Vehicles: err = %v, want *CantCastError", err)
	}
}

// TestTheIndomitableGraveyardCastSucceedsWithThreeTappedVehicles —
// the condition doesn't require Pirates specifically: three tapped
// Vehicles (uncrewed, so not even creatures) satisfy "Pirates and/or
// Vehicles" on their own.
func TestTheIndomitableGraveyardCastSucceedsWithThreeTappedVehicles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.SeedGraveyardCard(t, g, "The Indomitable", "Legendary Artifact — Vehicle", theIndomitableOracle)
	effectstest.B06AddMana(me, "U", "U", "C", "C")

	for i := 0; i < 3; i++ {
		v := effectstest.PushVehicleForTest(g, me.ID, fmt.Sprintf("Tapped Vehicle %d", i), "", 1, 1)
		for j := range g.Battlefield.Cards {
			if g.Battlefield.Cards[j].InstanceID == v {
				g.Battlefield.Cards[j].Tapped = true
			}
		}
	}

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast with three tapped Vehicles: %v", err)
	}
}

// TestTheIndomitableHandCastIsNeverGated — the printed clause only
// restricts the GRAVEYARD path; a normal hand cast never checks the
// board.
func TestTheIndomitableHandCastIsNeverGated(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	indom := effectstest.CastCatalogSpell(t, g, "The Indomitable", "Legendary Artifact — Vehicle", theIndomitableOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(indom) {
		t.Fatal("The Indomitable did not resolve onto the battlefield")
	}
}
