package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloudhoof Kirin — Legendary Creature — Kirin Spirit {3}{U}{U}, 4/4
// (Saviors of Kamigawa, #5601):
//
//	"Flying
//	 Whenever you cast a Spirit or Arcane spell, you may have target
//	 player mill X cards, where X is that spell's mana value."
//
// The cast event's Amount is the spell's mana value as cast (CR 202.3,
// with X chosen, CR 202.3e), read as the trigger resolves. The player is
// chosen as the trigger goes on the stack; the "may" is asked as it
// resolves (CR 603.5); a player gone by then is not milled (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "82d8732b-8246-4fbf-96aa-e192daf527f1",
		Name:            "Cloudhoof Kirin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Optional(effects.WheneverYouCast(effects.ChkB1SpiritOrArcane(), "Cloudhoof Kirin — target player mills X cards, X that spell's mana value",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player, ok := effects.FirstLegalPlayerTarget(ctx)
					if !ok || item.Trigger == nil {
						return nil
					}
					return effects.MillCards{Player: player, N: item.Trigger.Event.Amount}.Apply(ctx)
				}), "Cloudhoof Kirin — have target player mill cards equal to that spell's mana value?"),
				effects.TargetPlayer("target player")),
		},
	})
}
