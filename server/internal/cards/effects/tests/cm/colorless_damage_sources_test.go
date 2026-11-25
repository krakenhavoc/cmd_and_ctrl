package testscm

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// colorless_damage_sources_test.go — #6230: Spec.ColorlessDamageSources
// reaches the engine through game.CatalogColorlessDamageSources, so a
// card declaring Ghostly Flame's static lets a black creature's damage
// through protection from black.
func TestColorlessDamageSourcesSpecReachesTheEngine(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:               "test-6230-colorless-damage-sources",
		Name:                   "Test Flame",
		Completeness:           effects.CompletenessFull,
		ColorlessDamageSources: []string{"B", "R"},
	})
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := effectstest.B12Creature(g, opp.ID, "Pro-Black Knight", "Creature — Knight", 2, 2)
	shade := effectstest.B12Creature(g, me.ID, "Shade", "Creature — Shade", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case knight:
				g.Battlefield.Cards[i].Keywords = []string{"protection from black"}
			case shade:
				g.Battlefield.Cards[i].Colors = []string{"B"}
			}
		}
	})
	deal := func() int {
		g.WithWriteLock(func() {
			if err := g.DealDamageToCreatureForEffect(shade, knight, 1); err != nil {
				t.Fatalf("damage: %v", err)
			}
		})
		c, _ := effectstest.BattlefieldCard(g, knight)
		return c.DamageMarked
	}
	if got := deal(); got != 0 {
		t.Fatalf("protection from black let a black source's damage through without the static: %d", got)
	}
	effectstest.B12Push(g, me.ID, "Test Flame", "Enchantment", "test-6230-colorless-damage-sources", 0, 0)
	if got := deal(); got != 1 {
		t.Errorf("with the static the damage was still prevented: %d damage, want 1", got)
	}
}
