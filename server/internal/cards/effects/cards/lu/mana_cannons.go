package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mana Cannons — Enchantment {2}{R} (Dominaria United, S82 DMU batch 15):
//
//	"Whenever you cast a multicolored spell, this enchantment deals X damage
//	 to any target, where X is the number of colors that spell is."
//
// A cast trigger on the controller's own multicolored spells; it targets as
// it goes on the stack (CR 603.3d). X is the number of the spell's colors,
// read as the ability resolves from the spell as it last existed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "26d28b68-e03b-4367-ae4b-5e54f42f8a88",
		Name:         "Mana Cannons",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverYouCast(effects.Multicolored(),
					"Mana Cannons — X damage to any target, where X is the number of colors that spell is",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						spell, ok := g.LookupCardForEffect(ctx.Trigger().Event.CardID)
						if !ok {
							return nil
						}
						x := effects.ColorsOf(spell)
						t, ok := ctx.ClauseTarget(0)
						if !ok || x <= 0 {
							return nil
						}
						return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: x}.Apply(ctx)
					}),
				effects.TargetAny()),
		},
	})
}
