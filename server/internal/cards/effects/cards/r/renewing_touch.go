package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Renewing Touch — {G} Sorcery (Starter 1999, S82 OLD batch 702, #5620):
//
//	"Shuffle any number of target creature cards from your graveyard into your
//	 library."
//
// Perpetual Timepiece's shuffle (R38ShuffleTargetGraveyardCardsIntoLibrary)
// over creature cards only; the library is shuffled even with no target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1bd5fd97-f898-4e17-9e8f-5fa93b9eef71",
		Name:         "Renewing Touch",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("any number of target creature cards from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 0),
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.R38ShuffleTargetGraveyardCardsIntoLibrary(ctx.Game, item)
		},
	})
}
