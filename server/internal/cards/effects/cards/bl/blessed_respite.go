package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blessed Respite — Instant {1}{G} (Modern Horizons 2, #5501):
//
//	"Target player shuffles their graveyard into their library. Prevent all
//	 combat damage that would be dealt this turn."
//
// The graveyard goes into that player's library and the library is
// shuffled however many cards moved (Finale of Revelation's body, aimed at
// the target); the prevention is Fog's, for the rest of the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c888122-9d05-4641-931c-83f771ba9d09",
		Name:         "Blessed Respite",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      game.Purpose{Answers: game.AnswerPrevent},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetPlayer && ctx.IsTargetLegal(item.Targets[0]) {
				if err := effects.FinaleShuffleGraveyardIntoLibrary(ctx, item.Targets[0].ID); err != nil {
					return err
				}
			}
			return effects.PreventAllCombatDamageThisTurn{Label: "Blessed Respite: prevent combat damage"}.Apply(ctx)
		},
	})
}
