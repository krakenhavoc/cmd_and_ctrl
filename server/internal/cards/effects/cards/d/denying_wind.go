package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Denying Wind — Sorcery {7}{U}{U} (Prophecy, S82 OLD batch 673, #5617):
//
//	"Search target player's library for up to seven cards and exile them. Then
//	 that player shuffles."
//
// Nightmare Incursion's search: the caster chooses up to seven, the library
// searched and shuffled is the target's, any player may be the target. Fewer
// than seven cards in the library is fine, and so is choosing fewer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8daed4d8-1a43-4452-85dd-51b320714489",
		Name:         "Denying Wind",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			return effects.SearchLibrary{
				Player:       ctx.Controller(),
				LibraryOwner: item.Targets[0].ID,
				Dest:         game.ZoneExile,
				Limit:        7,
				UpTo:         true,
				Shuffle:      true,
				Source:       ctx.Source(),
				Reason:       "Denying Wind — search target player's library for up to seven cards and exile them",
			}.Apply(ctx)
		},
	})
}
