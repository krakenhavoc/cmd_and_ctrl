package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ember Gale — Sorcery {3}{R} (Shadowmoor, S82 OLD b530, #5591):
//
//	"Creatures target player controls can't block this turn. Ember Gale deals 1
//	 damage to each white and/or blue creature that player controls."
//
// Headliner Scarlett's scope for the first sentence (a rules effect, read live
// until cleanup, so a creature that player gets later can't block either), then
// 1 damage from the spell to each white or blue creature that player controls,
// dealt at once. With its one target illegal the spell does nothing (CR
// 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d93633d6-8433-43c8-9224-a25b411bfd96",
		Name:         "Ember Gale",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := effects.FirstLegalTargetPlayer(ctx)
			if player == [16]byte{} {
				return nil
			}
			return effects.OldB530CantBlockAndDamageWhiteAndBlue(ctx, player, "Ember Gale — creatures that player controls can't block this turn")
		},
	})
}
