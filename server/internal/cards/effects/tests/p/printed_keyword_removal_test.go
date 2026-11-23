package testsp

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// printed_keyword_removal_test.go — #4236: a catalogued creature's printed
// keywords are the object's own characteristics (CR 613.1), the baseline the
// layers start from, so a removal takes them away whenever it began. They
// used to be a layer-6 row at the creature's own timestamp, which put a
// keyword back after any older "loses flying".

const serraAngelTypeLine = "Creature — Angel"

// Zahid, a catalogued flier, enters before or after Gravity Sphere. Either
// way she loses flying: a static ability's effect is not tied to the
// objects there when it began (CR 613.7), and printed flying is no effect
// that could out-wait it. A flying grant newer than the Sphere gives it
// back, because the Sphere says "lose", not "can't have".
func TestGravitySphereRemovesACataloguedCreaturesPrintedFlyingInEitherOrder(t *testing.T) {
	for _, sphereFirst := range []bool{true, false} {
		name := "Sphere older"
		if !sphereFirst {
			name = "Sphere newer"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			var zahid uuid.UUID
			if sphereFirst {
				effectstest.PushCatalogPermanent(g, me.ID, "Gravity Sphere", "World Enchantment", effectstest.GravitySphereOracle, false)
				zahid = effectstest.PushCatalogPermanent(g, me.ID, "Zahid, Djinn of the Lamp", "Legendary Creature — Djinn", effectstest.ZahidOracle, false)
			} else {
				zahid = effectstest.PushCatalogPermanent(g, me.ID, "Zahid, Djinn of the Lamp", "Legendary Creature — Djinn", effectstest.ZahidOracle, false)
				if !effectstest.KrgHas(t, g, zahid, "flying") {
					t.Fatal("setup: Zahid has printed flying before the Sphere")
				}
				effectstest.PushCatalogPermanent(g, me.ID, "Gravity Sphere", "World Enchantment", effectstest.GravitySphereOracle, false)
			}
			if effectstest.KrgHas(t, g, zahid, "flying") {
				t.Error("Zahid kept printed flying under Gravity Sphere (stronger than printed)")
			}
			effectstest.KrgGrantUntilEOT(t, g, me.ID, zahid, "flying")
			if !effectstest.KrgHas(t, g, zahid, "flying") {
				t.Error("a flying grant newer than the Sphere did not give flying back")
			}
		})
	}
}

// Mystic Decree takes flying and islandwalk and leaves the rest of a
// catalogued creature's printed keywords.
func TestOlderMysticDecreeLeavesTheOtherPrintedKeywordsOfACataloguedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Mystic Decree", "World Enchantment", effectstest.MysticDecreeOracle, false)
	serra := effectstest.PushCatalogPermanent(g, me.ID, "Serra Angel", serraAngelTypeLine, effectstest.SerraAngelOracle, false)
	if effectstest.KrgHas(t, g, serra, "flying") {
		t.Error("Serra Angel kept flying under an older Mystic Decree")
	}
	if !effectstest.KrgHas(t, g, serra, "vigilance") {
		t.Error("Mystic Decree took vigilance away")
	}
}

// "Loses all abilities" empties the baseline like any other removal, and a
// grant that sorts after it survives (CR 613.6, 613.7): Zahid enchanted by
// Darksteel Mutation has no flying, and gets it back from a later grant.
func TestDarksteelMutationRemovesACataloguedCreaturesPrintedFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	zahid := effectstest.PushCatalogPermanent(g, me.ID, "Zahid, Djinn of the Lamp", "Legendary Creature — Djinn", effectstest.ZahidOracle, false)
	if !effectstest.KrgHas(t, g, zahid, "flying") {
		t.Fatal("setup: Zahid has printed flying")
	}
	effectstest.Enchant(t, g, "Darksteel Mutation", effectstest.DarksteelMutationOracle, zahid)
	if effectstest.KrgHas(t, g, zahid, "flying") {
		t.Error("Zahid kept flying under Darksteel Mutation")
	}
	if !effectstest.KrgHas(t, g, zahid, "indestructible") {
		t.Error("the Mutation's own indestructible is gone")
	}
	effectstest.KrgGrantUntilEOT(t, g, me.ID, zahid, "flying")
	if !effectstest.KrgHas(t, g, zahid, "flying") {
		t.Error("a flying grant newer than the Mutation did not give flying back")
	}
}

// The layered keywords are recomputed from the board on a restore: a
// catalogued creature under an older Sphere is still without flying, and
// a grant made afterwards is still a grant.
func TestAnOlderSphereStillTakesPrintedFlyingAfterARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Gravity Sphere", "World Enchantment", effectstest.GravitySphereOracle, false)
	zahid := effectstest.PushCatalogPermanent(g, me.ID, "Zahid, Djinn of the Lamp", "Legendary Creature — Djinn", effectstest.ZahidOracle, false)
	if effectstest.KrgHas(t, g, zahid, "flying") {
		t.Fatal("setup: Zahid has no flying under the Sphere")
	}
	r := effectstest.RestoreRoundTrip(t, g, true)
	if effectstest.KrgHas(t, r, zahid, "flying") {
		t.Error("Zahid has flying again after the restore")
	}
	effectstest.KrgGrantUntilEOT(t, r, me.ID, zahid, "flying")
	if !effectstest.KrgHas(t, r, zahid, "flying") {
		t.Error("a grant on the restored table did not give flying")
	}
}
