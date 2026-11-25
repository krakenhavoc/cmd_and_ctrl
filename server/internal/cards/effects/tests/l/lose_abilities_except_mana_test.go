package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lose_abilities_except_mana_test.go — #6246: "All lands lose all
// abilities except mana abilities." (Blood Sun; CR 613.1f, CR 605.1a),
// effects.LandsLoseAllAbilitiesExceptMana under a test key. The card
// ships with its OLD batch.

const (
	bloodSunTestKey      = "test-6246-blood-sun"
	exceptManaWilds      = "a75445d3-1303-4bb5-89ad-26ea93fecd48" // Evolving Wilds
	exceptManaCradle     = "7c427c3d-ecd8-45ef-bebd-8f10f4a311db" // Gaea's Cradle
	exceptManaKhalni     = "b2d5ba45-8674-4428-89db-c2bbbf0bf5c5" // Khalni Garden
	exceptManaLanternKey = "539f5396-d99a-417d-a84c-dff7930b5900" // Chromatic Lantern
)

func withBloodSun(t *testing.T) {
	t.Helper()
	prev := game.CatalogStaticAbilities
	row := []game.StaticAbility{effects.LandsLoseAllAbilitiesExceptMana()}
	game.CatalogStaticAbilities = func(k string) []game.StaticAbility {
		if k == bloodSunTestKey {
			return row
		}
		return prev(k)
	}
	t.Cleanup(func() { game.CatalogStaticAbilities = prev })
}

func exceptManaCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var c game.Card
	var ok bool
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok = g.LookupCardForEffect(id)
	})
	if !ok {
		t.Fatalf("card %v not found", id)
	}
	return c
}

// The acceptance: a fetchland loses its sacrifice ability, Gaea's Cradle
// keeps its mana ability, and a land with an enters trigger doesn't
// trigger. Without Blood Sun the fetchland's ability is there.
func TestBloodSunLandsKeepOnlyManaAbilities(t *testing.T) {
	withBloodSun(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wilds := effectstest.PushCatalogPermanent(g, me.ID, "Evolving Wilds", "Land", exceptManaWilds, false)
	cradle := effectstest.PushCatalogPermanent(g, me.ID, "Gaea's Cradle", "Legendary Land", exceptManaCradle, false)
	if n := len(game.ActivatedAbilitiesForCard(exceptManaCard(t, g, wilds))); n == 0 {
		t.Fatal("setup: Evolving Wilds has no activated ability without Blood Sun")
	}

	effectstest.PushCatalogPermanent(g, me.ID, "Blood Sun", "Enchantment", bloodSunTestKey, false)
	w := exceptManaCard(t, g, wilds)
	if n := len(game.ActivatedAbilitiesForCard(w)); n != 0 {
		t.Errorf("Evolving Wilds has %d activated abilities under Blood Sun, want 0", n)
	}
	if err := g.ActivateCatalogAbility(me.ID, wilds, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("Evolving Wilds' sacrifice ability activated under Blood Sun")
	}
	if n := len(game.ManaAbilitiesForCard(exceptManaCard(t, g, cradle))); n == 0 {
		t.Error("Gaea's Cradle lost its mana ability under Blood Sun")
	}

	khalni := effectstest.PushCatalogPermanent(g, me.ID, "Khalni Garden", "Land", exceptManaKhalni, false)
	if n := len(game.TriggersForCard(exceptManaCard(t, g, khalni))); n != 0 {
		t.Errorf("Khalni Garden has %d triggered abilities under Blood Sun, want 0", n)
	}
	if n := len(game.ManaAbilitiesForCard(exceptManaCard(t, g, khalni))); n == 0 {
		t.Error("Khalni Garden lost its mana ability under Blood Sun")
	}
}

// A basic land keeps the mana ability of its type, and a land granted a
// mana ability (Chromatic Lantern) keeps that too: it is a mana ability.
func TestBloodSunKeepsIntrinsicAndGrantedManaAbilities(t *testing.T) {
	withBloodSun(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.PushTypedCard(g, me.ID, "Forest", "Basic Land — Forest", "")
	wilds := effectstest.PushCatalogPermanent(g, me.ID, "Evolving Wilds", "Land", exceptManaWilds, false)
	effectstest.PushCatalogPermanent(g, me.ID, "Chromatic Lantern", "Artifact", exceptManaLanternKey, false)
	if n := len(game.ManaAbilitiesForCard(exceptManaCard(t, g, wilds))); n == 0 {
		t.Fatal("setup: Chromatic Lantern gives Evolving Wilds no mana ability")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Blood Sun", "Enchantment", bloodSunTestKey, false)
	if n := len(game.ManaAbilitiesForCard(exceptManaCard(t, g, forest))); n == 0 {
		t.Error("a Forest lost its intrinsic mana ability under Blood Sun")
	}
	w := exceptManaCard(t, g, wilds)
	if n := len(game.ManaAbilitiesForCard(w)); n == 0 {
		t.Error("Evolving Wilds lost Chromatic Lantern's granted mana ability under Blood Sun")
	}
	if n := len(game.ActivatedAbilitiesForCard(w)); n != 0 {
		t.Errorf("Evolving Wilds kept %d non-mana activated abilities", n)
	}
}
