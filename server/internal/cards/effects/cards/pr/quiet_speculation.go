package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quiet Speculation — Sorcery {1}{U} (Eternal Masters, S82 OLD, #5542):
//
//	"Search target player's library for up to three cards with flashback and
//	 put them into that player's graveyard. Then the player shuffles."
//
// Jester's Cap's search of another player's library by the caster, here for up
// to three cards that have flashback (a stated quality, so the search may find
// fewer, CR 701.23b), into the graveyard of the library's owner, who shuffles
// afterwards. Any player may be the target, the caster included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "938f48a9-766e-41d3-a425-9c3a6adf675b",
		Name:         "Quiet Speculation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return effects.SearchLibrary{
				Player:       ctx.Controller(),
				LibraryOwner: item.Targets[0].ID,
				Predicate:    game.HasFlashback,
				Dest:         game.ZoneGraveyard,
				Limit:        3,
				Shuffle:      true,
				Reason:       "Quiet Speculation — search target player's library for up to three cards with flashback and put them into their graveyard",
			}.Apply(ctx)
		},
	})
}
