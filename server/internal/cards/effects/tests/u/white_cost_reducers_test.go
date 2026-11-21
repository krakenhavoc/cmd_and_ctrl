package testsu

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	pearlMedallionOracle = "19015380-1332-4960-8cc6-0732009525a2"
	theWindCrystalOracle = "240f0835-36af-4ad8-9336-d6d3d816d293"
)

// TestPearlMedallionDiscountsYourWhiteSpellsOnly — "WHITE spells YOU
// cast": a white spell of any type, generic mana only, and nobody
// else's.
func TestPearlMedallionDiscountsYourWhiteSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Pearl Medallion", "Artifact", pearlMedallionOracle, false)

	if got := effectstest.PriceInHand(t, g, me, "White Instant", "Instant", "{1}{W}"); got != 1 {
		t.Errorf("own white instant: %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "White Bear", "Creature — Bear", "{2}{W}"); got != 2 {
		t.Errorf("own white creature: %d, want 2", got)
	}
	// CR 601.2f: no generic pip, nothing to reduce.
	if got := effectstest.PriceInHand(t, g, me, "Swords", "Instant", "{W}"); got != 1 {
		t.Errorf("own {W} spell: %d, want 1 (untouched)", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Red Instant", "Instant", "{1}{R}"); got != 2 {
		t.Errorf("own red spell: %d, want 2 (untouched)", got)
	}
	if got := effectstest.PriceInHand(t, g, them, "Their White Instant", "Instant", "{1}{W}"); got != 2 {
		t.Errorf("opponent's white spell: %d, want 2 (undiscounted)", got)
	}
}

// TestTheWindCrystalDiscountsYourWhiteSpells is the first line.
func TestTheWindCrystalDiscountsYourWhiteSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "The Wind Crystal", "Legendary Artifact", theWindCrystalOracle, false)

	if got := effectstest.PriceInHand(t, g, me, "White Instant", "Instant", "{1}{W}"); got != 1 {
		t.Errorf("own white spell: %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Blue Instant", "Instant", "{1}{U}"); got != 2 {
		t.Errorf("own blue spell: %d, want 2 (untouched)", got)
	}
	if got := effectstest.PriceInHand(t, g, them, "Their White Instant", "Instant", "{1}{W}"); got != 2 {
		t.Errorf("opponent's white spell: %d, want 2 (undiscounted)", got)
	}
}

// TestTheWindCrystalDoublesOnlyYourLifeGain is the second line: a
// catalog gain of three is six for the Crystal's controller and three
// for anyone else.
func TestTheWindCrystalDoublesOnlyYourLifeGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "The Wind Crystal", "Legendary Artifact", theWindCrystalOracle, false)
	mine, theirs := me.Life, opp.Life

	effectstest.B11Helix(t, g, opp.ID)
	if me.Life != mine+6 {
		t.Errorf("Lightning Helix's gain of 3 under the Crystal: %d → %d, want %d", mine, me.Life, mine+6)
	}

	g.WithWriteLock(func() {
		if err := (effects.GainLife{Player: opp.ID, Amount: 3}).Apply(effects.NewContext(g, &game.StackItem{Controller: opp.ID})); err != nil {
			t.Fatalf("GainLife: %v", err)
		}
	})
	if opp.Life != theirs-3+3 {
		t.Errorf("an opponent's gain of 3 is not doubled: %d → %d, want %d", theirs-3, opp.Life, theirs)
	}
}

// TestTheWindCrystalGrantsFlyingAndLifelinkToYourCreatures is the
// third line, through a real activation: {4}{W}{W} and {T}, and the
// grant lands on the creatures you control as it resolves (CR 611.2c)
// — not on an opponent's.
func TestTheWindCrystalGrantsFlyingAndLifelinkToYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	crystal := effectstest.PushCatalogPermanent(g, me.ID, "The Wind Crystal", "Legendary Artifact", theWindCrystalOracle, false)
	mine := effectstest.PushPermanentForTest(g, me.ID, "My Bear", "", "Creature — Bear")
	theirs := effectstest.PushPermanentForTest(g, opp.ID, "Their Bear", "", "Creature — Bear")

	effectstest.FloatForTest(g, me, "CCCCWW")
	if err := g.ActivateCatalogAbility(me.ID, crystal, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if c, _ := effectstest.BattlefieldCard(g, crystal); !c.Tapped {
		t.Error("the ability's {T} did not tap the Crystal")
	}
	effectstest.PassPriorityAroundTable(t, g)

	for _, kw := range []string{"flying", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, mine, kw) {
			t.Errorf("your creature did not gain %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("an opponent's creature gained %s", kw)
		}
	}
}
