package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daring Piracy — Enchantment {2}{R} (Jumpstart 2022):
//
//	"At the beginning of combat on your turn, create a 1/1 red Pirate creature
//	 token with menace and haste. Exile it at the beginning of the next end
//	 step."
//
// Goblin Rabblemaster's beginning-of-combat token with Splinter Twin's
// delayed exile: the token made by this resolution is found by its creation
// and one delayed trigger (CR 603.7) exiles it at the next end step, this
// turn's. A token that has already left is not exiled again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae74efc6-4649-4783-a3a5-7a34d4667f64",
		Name:         "Daring Piracy",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Daring Piracy — create a 1/1 red Pirate with menace and haste; exile it at the beginning of the next end step",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					cursor := effects.B25LastEventSeq(g)
					if err := (effects.CreateToken{Template: effects.TokenCard("1/1 red Pirate with menace and haste"), N: 1}).Apply(ctx); err != nil {
						return err
					}
					tokens := effects.B27TokensCreatedByAfter(g, item.Controller, cursor)
					if len(tokens) == 0 {
						return nil
					}
					return effects.ScheduleDelayedTrigger{
						At:         game.StepEnd,
						Controller: item.Controller,
						Label:      "Daring Piracy — exile the Pirate token",
						Cards:      tokens,
						Body:       effects.ExileListedCardsBody,
					}.Apply(ctx)
				}),
		},
	})
}
