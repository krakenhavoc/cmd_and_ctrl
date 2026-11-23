package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hullbreacher — Creature — Merfolk Pirate {2}{U}, 3/2 (Mystery Booster 2):
//
//	"Flash
//	 If an opponent would draw a card except the first one they draw in each of
//	 their draw steps, instead you create a Treasure token."
//
// Notion Thief's draw replacement (the same AppliesTo), with a Treasure for
// the Hullbreacher's controller in place of the draw: the opponent's draw is
// cancelled, so it is not a draw (CR 614.6) and no "whenever you draw" fires,
// and the rest of a multi-card draw goes on, each card asked on its own
// (CR 121.2). Flash is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2f533667-e29b-4bec-897a-e7a9eee08314",
		Name:            "Hullbreacher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Replacements: []game.ReplacementEffect{{
			Watches:     []game.EventKind{game.EventDrawCard},
			AppliesTo:   effects.S82GapB24OpponentDrawsExceptTheFirst,
			DrawInstead: game.RegisterDrawInstead("hullbreacher-treasure", effects.S82GapB24TreasureInstead),
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Hullbreacher: you create a Treasure token instead",
		}},
	})
}
