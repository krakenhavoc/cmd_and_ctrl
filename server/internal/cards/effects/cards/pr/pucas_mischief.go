package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Puca's Mischief — Enchantment {3}{U} (Secret Lair Drop, #5462):
//
//	"At the beginning of your upkeep, you may exchange control of target
//	 nonland permanent you control and target nonland permanent an
//	 opponent controls with equal or lesser mana value."
//
// Switcheroo's exchange (CR 701.12) as a "you may" upkeep trigger over two
// targets. The second target is related to the first: its mana value must be
// equal to or less than the first's (ManaValueNoGreaterThanTarget), judged when
// the trigger is put on the stack and again as it resolves. The "may" is asked
// as the trigger resolves, and the exchange is all or nothing (CR 701.12a): if
// either permanent is illegal by then, nothing changes hands. Neither half has
// a duration (CR 611.2a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a8e6655c-e63f-4a60-98f6-9d5bf8631488",
		Name:         "Puca's Mischief",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.AtYourUpkeep("Puca's Mischief — you may exchange control of target nonland permanent you control and target nonland permanent an opponent controls with equal or lesser mana value",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						mine, ok := ctx.ClauseTarget(0)
						if !ok {
							return nil
						}
						theirs, ok := ctx.ClauseTarget(1)
						if !ok {
							return nil
						}
						return effects.ExchangeControl{A: mine.ID, B: theirs.ID, Label: "Puca's Mischief — exchange control"}.Apply(ctx)
					}),
				effects.Clauses(
					effects.TargetPermanent("target nonland permanent you control", effects.Nonland(), effects.YouControl()),
					effects.ManaValueNoGreaterThanTarget(
						effects.TargetPermanent("target nonland permanent an opponent controls with equal or lesser mana value", effects.Nonland(), effects.OpponentControls()), 0),
				)),
				"Puca's Mischief — exchange control?"),
		},
	})
}
