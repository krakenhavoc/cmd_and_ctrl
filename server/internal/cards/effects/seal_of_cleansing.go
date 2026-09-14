package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seal of Cleansing — Enchantment {1}{W} (EDHREC rank 3284):
//
//	"Sacrifice this enchantment: Destroy target artifact or
//	 enchantment."
//
// Disenchant on a permanent, paid for up front and cashed at
// instant speed later. The ability is Aura of Silence's second
// line: a sacrifice-this activation with no mana and no tap, the
// target picked at announce and re-checked at resolution (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a75dbe70-7e3e-446f-9a76-9fbb414f2e7c",
		Name:         "Seal of Cleansing",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this enchantment: Destroy target artifact or enchantment.",
			Cost:    SacrificeThis(),
			Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return DestroyTarget{Target: t.ID}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
