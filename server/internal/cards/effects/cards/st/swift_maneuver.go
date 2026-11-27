package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swift Maneuver — Instant {1}{W} (Coldsnap):
//
//	"Prevent the next 2 damage that would be dealt to any target this turn.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// The shield is a CR 615.7 prevention of the next 2 damage dealt to the
// target (player, creature or planeswalker) this turn; the draw is a delayed
// trigger set up as the spell resolves (CR 603.7) and still happens when the
// target is gone but the spell resolved.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7b5fb4b-8f3c-4143-9fb0-2e2544612528",
		Name:         "Swift Maneuver",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPrevent, Draws: 1},
		Targets:      effects.TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.PreventNextDamageToTheTarget("Swift Maneuver", 2)(ctx.Game, ctx.Item); err != nil {
				return err
			}
			return effects.DrawACardAtTheNextUpkeep("Swift Maneuver — draw a card").Apply(ctx)
		},
	})
}
