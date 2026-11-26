package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conjured Currency — Enchantment {5}{U} (Return to Ravnica, #5566):
//
//	"At the beginning of your upkeep, you may exchange control of this
//	 enchantment and target permanent you neither own nor control."
//
// Puca's Mischief's "you may exchange control" upkeep trigger, with the
// enchantment itself as one half. The target is a permanent whose owner and
// controller are both someone else, judged on putting the trigger on the stack
// and again as it resolves (CR 608.2b). The "may" is asked on resolution. The
// exchange is all or nothing (CR 701.12a) and does nothing if the enchantment
// has left or come back as a new object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "427ba80d-cfdf-4651-a1eb-89a2125a9217",
		Name:         "Conjured Currency",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.AtYourUpkeep("Conjured Currency — you may exchange control of this enchantment and target permanent you neither own nor control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok || !effects.SourceIsStillThisPermanent(g, item) {
							return nil
						}
						return effects.ExchangeControl{A: item.SourceCardID, B: id, Label: "Conjured Currency — exchange control"}.Apply(ctx)
					}),
				effects.TargetPermanent("target permanent you neither own nor control",
					func(_ *game.Game, caster uuid.UUID, c game.Card) bool {
						return c.Owner != caster && c.Controller != caster
					})),
				"Conjured Currency — exchange control?"),
		},
	})
}
