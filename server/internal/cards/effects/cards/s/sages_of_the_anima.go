package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sages of the Anima — Creature — Elf Wizard {3}{G}{U}, 3/4 (Alara Reborn,
// #5586):
//
//	"If you would draw a card, instead reveal the top three cards of your
//	 library. Put all creature cards revealed this way into your hand and the
//	 rest on the bottom of your library in any order."
//
// Underrealm Lich's replacement of every draw you make: the draw is cancelled
// (nothing is "drawn", CR 614.6), the top three are revealed to everyone, every
// creature card among them goes to the hand and the rest are ordered by you on
// the bottom. A "draw three" asks three times (CR 121.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6c008c7-6c87-4677-ba0e-6d95f10d10bc",
		Name:         "Sages of the Anima",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventDraw && ev.DrawCount > 0 && src != nil && ev.DrawPlayer == src.Controller
			},
			DrawInstead: game.RegisterDrawInstead("sages-of-the-anima", effects.S82OldB501SagesInstead),
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Sages of the Anima: reveal the top three, put the creature cards into your hand and the rest on the bottom",
		}},
	})
}
