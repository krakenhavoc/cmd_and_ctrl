package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Honden of Infinite Rage — Legendary Enchantment — Shrine {2}{R}
// (Champions of Kamigawa, #5603):
//
//	"At the beginning of your upkeep, Honden of Infinite Rage deals
//	 damage to any target equal to the number of Shrines you control."
//
// The target is chosen as the trigger goes on the stack and the Shrines
// are counted as it resolves (CR 608.2h), so a Shrine destroyed in
// response no longer counts. The Honden is the source of the damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0cb76b04-d9c3-46fc-a3f2-5f0cb0d91761",
		Name:         "Honden of Infinite Rage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(chkB1Honden("Honden of Infinite Rage — deals damage to any target equal to the number of Shrines you control",
				func(g *game.Game, item *game.StackItem, shrines int) error {
					ctx := NewContext(g, item)
					t, ok := firstLegalTarget(ctx)
					if !ok {
						return nil
					}
					return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: shrines}.Apply(ctx)
				}), TargetAny()),
		},
	})
}
