package testscm

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// colorless_damage_sources_test.go — #6230: effects.ColorlessDamageSources
// (Ghostly Flame) under a test key; the card ships with its OLD batch.

const colorlessSourcesTestKey = "test-6230-ghostly-flame"

func TestColorlessDamageSourcesHelperLetsBlackDamageThroughProtection(t *testing.T) {
	prev := game.CatalogStaticAbilities
	row := []game.StaticAbility{effects.ColorlessDamageSources("BR", "Black and/or red permanents and spells are colorless sources of damage.")}
	game.CatalogStaticAbilities = func(k string) []game.StaticAbility {
		if k == colorlessSourcesTestKey {
			return row
		}
		return prev(k)
	}
	t.Cleanup(func() { game.CatalogStaticAbilities = prev })

	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.PushTypedCard(g, me.ID, "Black Knight", "Creature — Knight", "{B}{B}")
	victim := effectstest.PushTypedCard(g, opp.ID, "White Knight", "Creature — Knight", "{W}{W}")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			switch c.InstanceID {
			case src:
				c.Colors = []string{"B"}
			case victim:
				c.Keywords = []string{"protection from black"}
				c.Toughness = 5
			}
		}
	})
	effectstest.PushCatalogPermanent(g, me.ID, "Ghostly Flame", "Enchantment", colorlessSourcesTestKey, false)
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if err := g.DealDamageToCreatureForEffect(src, victim, 2); err != nil {
			t.Fatal(err)
		}
	})
	var marked int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == victim {
				marked = c.DamageMarked
			}
		}
	})
	if marked != 2 {
		t.Errorf("protection from black prevented %d of 2: under Ghostly Flame a black source is colourless", 2-marked)
	}
}
