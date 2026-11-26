package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pieces of the Puzzle — Sorcery {2}{U} (Shadows over Innistrad, S82 OLD):
//
//	"Reveal the top five cards of your library. Put up to two instant and/or
//	 sorcery cards from among them into your hand and the rest into your
//	 graveyard."
//
// Dihada's reveal-and-take with a cap of two matching cards and the rest to
// the graveyard; the five are public (CR 701.20a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b599b4d-95bc-4ed2-a167-84a697944747",
		Name:         "Pieces of the Puzzle",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			revealed := ctx.Game.RevealTopOfLibraryForEffect(player, ctx.Source(), 5,
				"Pieces of the Puzzle — reveal the top five cards of your library")
			if len(revealed) == 0 {
				return nil
			}
			return effects.TakeFromLibraryToHand{
				Player:   player,
				Cards:    revealed,
				Match:    effects.Or(effects.Instant(), effects.Sorcery()),
				Max:      2,
				Optional: true,
				Label:    "Pieces of the Puzzle — put up to two instant and/or sorcery cards from among them into your hand",
				Key:      effects.TakeRestGraveyard,
			}.Apply(ctx)
		},
	})
}
