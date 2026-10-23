package effects

import (
	"testing"

	"github.com/google/uuid"
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
			g := newCatalogGame(t)
			me := g.Seats[0]
			var zahid uuid.UUID
			if sphereFirst {
				pushCatalogPermanent(g, me.ID, "Gravity Sphere", "World Enchantment", gravitySphereOracle, false)
				zahid = pushCatalogPermanent(g, me.ID, "Zahid, Djinn of the Lamp", "Legendary Creature — Djinn", zahidOracle, false)
			} else {
				zahid = pushCatalogPermanent(g, me.ID, "Zahid, Djinn of the Lamp", "Legendary Creature — Djinn", zahidOracle, false)
				if !krgHas(t, g, zahid, "flying") {
					t.Fatal("setup: Zahid has printed flying before the Sphere")
				}
				pushCatalogPermanent(g, me.ID, "Gravity Sphere", "World Enchantment", gravitySphereOracle, false)
			}
			if krgHas(t, g, zahid, "flying") {
				t.Error("Zahid kept printed flying under Gravity Sphere (stronger than printed)")
			}
			krgGrantUntilEOT(t, g, me.ID, zahid, "flying")
			if !krgHas(t, g, zahid, "flying") {
				t.Error("a flying grant newer than the Sphere did not give flying back")
			}
		})
	}
}

// Mystic Decree takes flying and islandwalk and leaves the rest of a
// catalogued creature's printed keywords.
func TestOlderMysticDecreeLeavesTheOtherPrintedKeywordsOfACataloguedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Mystic Decree", "World Enchantment", mysticDecreeOracle, false)
	serra := pushCatalogPermanent(g, me.ID, "Serra Angel", serraAngelTypeLine, serraAngelOracle, false)
	if krgHas(t, g, serra, "flying") {
		t.Error("Serra Angel kept flying under an older Mystic Decree")
	}
	if !krgHas(t, g, serra, "vigilance") {
		t.Error("Mystic Decree took vigilance away")
	}
}

// "Loses all abilities" empties the baseline like any other removal, and a
// grant that sorts after it survives (CR 613.6, 613.7): Zahid enchanted by
// Darksteel Mutation has no flying, and gets it back from a later grant.
func TestDarksteelMutationRemovesACataloguedCreaturesPrintedFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	zahid := pushCatalogPermanent(g, me.ID, "Zahid, Djinn of the Lamp", "Legendary Creature — Djinn", zahidOracle, false)
	if !krgHas(t, g, zahid, "flying") {
		t.Fatal("setup: Zahid has printed flying")
	}
	enchant(t, g, "Darksteel Mutation", darksteelMutationOracle, zahid)
	if krgHas(t, g, zahid, "flying") {
		t.Error("Zahid kept flying under Darksteel Mutation")
	}
	if !krgHas(t, g, zahid, "indestructible") {
		t.Error("the Mutation's own indestructible is gone")
	}
	krgGrantUntilEOT(t, g, me.ID, zahid, "flying")
	if !krgHas(t, g, zahid, "flying") {
		t.Error("a flying grant newer than the Mutation did not give flying back")
	}
}

// The layered keywords are recomputed from the board on a restore: a
// catalogued creature under an older Sphere is still without flying, and
// a grant made afterwards is still a grant.
func TestAnOlderSphereStillTakesPrintedFlyingAfterARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Gravity Sphere", "World Enchantment", gravitySphereOracle, false)
	zahid := pushCatalogPermanent(g, me.ID, "Zahid, Djinn of the Lamp", "Legendary Creature — Djinn", zahidOracle, false)
	if krgHas(t, g, zahid, "flying") {
		t.Fatal("setup: Zahid has no flying under the Sphere")
	}
	r := restoreRoundTrip(t, g, true)
	if krgHas(t, r, zahid, "flying") {
		t.Error("Zahid has flying again after the restore")
	}
	krgGrantUntilEOT(t, r, me.ID, zahid, "flying")
	if !krgHas(t, r, zahid, "flying") {
		t.Error("a grant on the restored table did not give flying")
	}
}
