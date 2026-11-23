package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Caress of Phyrexia — Sorcery {3}{B}{B}:
//
//	"Target player draws three cards, loses 3 life, and gets three poison
//	 counters."
//
// Sign in Blood's draw and life loss on the chosen player, then three
// poison counters through the CR 614 player-counter window, placed by the
// spell's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8253fcac-0c07-41e8-afa4-34021f9ef6f4",
		Name:         "Caress of Phyrexia",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 3, LifeLoss: 3}),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			target := item.Targets[0].ID
			if err := (effects.DrawCards{Player: target, N: 3}).Apply(ctx); err != nil {
				return err
			}
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), target, -3); err != nil {
				return err
			}
			return ctx.Game.AddPlayerCounterByForEffect(item.Controller, target, game.CounterPoison, 3)
		},
	})
}
