package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Budding Insurgent — Creature — Dryad Scout {2}{G}, 3/3:
//
//	"Vigilance
//	 Sacrifice this creature: Destroy target artifact or enchantment.
//	 If that permanent was a legendary enchantment, draw a card.
//	 Activate only as a sorcery."
//
// Sacrificing is a cost, so the Insurgent is gone before the ability
// resolves. Whether the target was a legendary enchantment is read
// before it is destroyed, so an indestructible one that survives still
// draws the card, as the rulings have it. If the target is gone or
// illegal at resolution the whole ability is skipped (CR 608.2b), draw
// included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "723669e0-4a04-45ec-8c76-b1a369aa6892",
		Name:            "Budding Insurgent",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:        "Sacrifice this creature: Destroy target artifact or enchantment. If that permanent was a legendary enchantment, draw a card. Activate only as a sorcery.",
			Cost:         SacrificeThis(),
			SorcerySpeed: true,
			Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				t := ts[0]
				c, ok := g.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				wasLegendaryEnchantment := c.IsLegendary() && c.IsEnchantment()
				// The draw runs in the destroy's continuation, so a
				// paused replacement window (CR 614) lands first.
				return g.DestroyPermanentsThenForEffect([]uuid.UUID{t.ID}, func(*game.Game, []uuid.UUID) error {
					if !wasLegendaryEnchantment {
						return nil
					}
					return DrawCards{N: 1}.Apply(ctx)
				})
			},
		}},
	})
}
