package effects

import (
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// magus_of_the_moon_test.go is the catalog proof for the two rules
// that held this card out of the catalog for four batches: CR 305.7
// (a basic-land-type set removes the land's OWN abilities, in layer
// 4, and gives it the type's mana ability) and CR 613.6 (a silenced
// source keeps applying in the layers that already ran).
//
// The engine contract for both is in game/layer_dependency_test.go.

// The headline sentence: a nonbasic land is a Mountain, taps for {R}
// and nothing else, and has lost the ability its own rules text
// printed. A basic land is untouched.
func TestMagusOfTheMoonMakesNonbasicLandsMountains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	grove := seedLand(g, me.ID, "Sunpetal Grove", "Land", sunpetalGroveOracle)
	basic := seedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	b31Push(g, me.ID, "Magus of the Moon", "Creature — Human Wizard", b31MagusOfTheMoonOracle, "{2}{R}", 2, 2, "R")

	if sub := effectiveSubtypes(t, g, grove); !slices.Contains(sub, "Mountain") || slices.Contains(sub, "Forest") {
		t.Errorf("Sunpetal Grove subtypes = %v, want exactly Mountain", sub)
	}
	if got := manaAbilityProduced(t, g, grove); len(got) != 1 || got[0] != "{R}" {
		t.Errorf("Sunpetal Grove produces %v, want only {R}: CR 305.7 removes its own ability and gives it the Mountain's", got)
	}
	if c := layeredCard(t, g, grove); !c.HasLostAllAbilities() {
		t.Error("the Grove kept its own rules text under CR 305.7")
	}
	if sub := effectiveSubtypes(t, g, basic); !slices.Contains(sub, "Forest") {
		t.Errorf("a BASIC land is not nonbasic (CR 205.4c): subtypes = %v", sub)
	}
	if got := manaAbilityProduced(t, g, basic); len(got) != 1 || got[0] != "{G}" {
		t.Errorf("the basic Forest produces %v, want {G}", got)
	}
}

// Magus of the Moon's 2021-03-19 ruling, as a catalog test: "If Magus
// of the Moon loses its abilities, it continues to turn nonbasic
// lands into Mountains." Kenrith's Transformation removes them in
// layer 6; the Mountains were made in layer 4 (CR 613.6).
//
// This is the assertion that fails with #669 backed out — the old
// recompute held a silenced source out of the GATHER, so the Magus
// contributed to no layer at all and the Grove went back to being a
// Sunpetal Grove.
func TestASilencedMagusOfTheMoonKeepsMakingMountains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	grove := seedLand(g, me.ID, "Sunpetal Grove", "Land", sunpetalGroveOracle)
	magus := b31Push(g, me.ID, "Magus of the Moon", "Creature — Human Wizard", b31MagusOfTheMoonOracle, "{2}{R}", 2, 2, "R")
	enchant(t, g, "Kenrith's Transformation", kenrithTransformOracle, magus)
	settle(t, g)

	if c := layeredCard(t, g, magus); !c.HasLostAllAbilities() {
		t.Fatal("fixture is wrong: Kenrith's Transformation did not silence the Magus")
	}
	if sub := effectiveSubtypes(t, g, grove); !slices.Contains(sub, "Mountain") {
		t.Errorf("Sunpetal Grove subtypes = %v, want Mountain: a silenced Magus keeps making them (CR 613.6)", sub)
	}
	if got := manaAbilityProduced(t, g, grove); len(got) != 1 || got[0] != "{R}" {
		t.Errorf("Sunpetal Grove produces %v, want only {R}", got)
	}
}

// CR 305.7's last sentence on the card that is not an Aura: the loss
// is the land's own rules text and nothing else, so a keyword another
// effect granted in layer 6 survives whenever it was granted.
func TestMagusOfTheMoonKeepsAnAbilityAnotherEffectGranted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	grove := seedLand(g, me.ID, "Sunpetal Grove", "Land", sunpetalGroveOracle)
	b31Push(g, me.ID, "Magus of the Moon", "Creature — Human Wizard", b31MagusOfTheMoonOracle, "{2}{R}", 2, 2, "R")
	castModal(t, g, "Boros Charm", "Instant", borosCharmOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)

	if c := layeredCard(t, g, grove); !game.HasKeyword(&c, "indestructible") {
		t.Errorf("the Mountain'd Grove lost Boros Charm's indestructible (abilities %v)", c.Effective().Abilities)
	}
}

// Urborg and the Magus are not a CR 613.8 dependency — "each land"
// applies to the same set whether or not the Magus has run, and "add
// Swamp" is the same instruction either way — so CR 613.7 timestamp
// order decides them, and the answer is the paper one on both sides.
//
// The Magus is itself the reason this is worth two subtests: Urborg
// is a nonbasic land, so the Magus silences it in LAYER 4, and
// whether that lands before or after Urborg's own effect is exactly
// what the timestamps say.
func TestMagusOfTheMoonAndUrborgSettleByTimestamp(t *testing.T) {
	// Urborg first: its Swamps are already on every land when the
	// Magus arrives. The Magus then overwrites the nonbasic ones and
	// leaves the basics alone, so a basic Forest keeps its Swamp.
	t.Run("Urborg older", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		grove := seedLand(g, me.ID, "Sunpetal Grove", "Land", sunpetalGroveOracle)
		forest := seedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
		seedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", urborgOracle)
		b31Push(g, me.ID, "Magus of the Moon", "Creature — Human Wizard", b31MagusOfTheMoonOracle, "{2}{R}", 2, 2, "R")

		sub := effectiveSubtypes(t, g, grove)
		if !slices.Contains(sub, "Mountain") || slices.Contains(sub, "Swamp") {
			t.Errorf("Grove subtypes = %v, want Mountain only: the newer Magus replaced the Swamp", sub)
		}
		if sub := effectiveSubtypes(t, g, forest); !slices.Contains(sub, "Swamp") {
			t.Errorf("basic Forest subtypes = %v, want the Swamp Urborg gave it", sub)
		}
	})
	// Magus first: Urborg is a nonbasic land, so the Magus takes its
	// abilities away in layer 4 BEFORE Urborg's own later timestamp
	// comes up, and Urborg does nothing at all. Nothing is a Swamp.
	t.Run("Magus older", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		grove := seedLand(g, me.ID, "Sunpetal Grove", "Land", sunpetalGroveOracle)
		forest := seedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
		b31Push(g, me.ID, "Magus of the Moon", "Creature — Human Wizard", b31MagusOfTheMoonOracle, "{2}{R}", 2, 2, "R")
		urborg := seedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", urborgOracle)

		if c := layeredCard(t, g, urborg); !c.HasLostAllAbilities() {
			t.Fatal("the Magus did not silence Urborg (CR 305.7 removes a nonbasic land's own text)")
		}
		if sub := effectiveSubtypes(t, g, forest); slices.Contains(sub, "Swamp") {
			t.Errorf("basic Forest subtypes = %v, want no Swamp: Urborg was silenced before its own effect applied", sub)
		}
		if sub := effectiveSubtypes(t, g, grove); slices.Contains(sub, "Swamp") {
			t.Errorf("Grove subtypes = %v, want no Swamp", sub)
		}
	})
}
