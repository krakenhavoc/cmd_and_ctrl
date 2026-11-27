package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tomorrow, Azami's Familiar — Legendary Creature — Spirit {5}{U}, 1/5
// (Betrayers of Kamigawa, #5602):
//
//	"If you would draw a card, look at the top three cards of your library
//	 instead. Put one of those cards into your hand and the rest on the
//	 bottom of your library in any order."
//
// Underrealm Lich's draw replacement (DrawInstead): mandatory, replacing each
// of your draws, a draw step's and an effect's alike. The draw is cancelled
// (nothing is drawn, so "whenever you draw" stays quiet, CR 614.6); the body
// looks at the top three (only you see them), asks which one to take, puts the
// rest on the bottom in the order you name and then runs the rest of the draw
// instruction, so "draw three" asks three times (CR 121.6b). A draw a doubler
// made into two is two draws and the replacement runs for each.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51d517c9-2812-44ce-ab4d-e5422b5ecf6c",
		Name:         "Tomorrow, Azami's Familiar",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventDraw && ev.DrawCount > 0 && src != nil && ev.DrawPlayer == src.Controller
			},
			DrawInstead: game.RegisterDrawInstead("s82-old-b600-tomorrow", effects.DrawInsteadLookTopThreeTakeOneBottomRest),
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Tomorrow, Azami's Familiar: look at the top three, put one into your hand and the rest on the bottom of your library",
		}},
	})
}
