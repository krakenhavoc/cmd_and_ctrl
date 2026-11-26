package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chancellor of the Dross — Creature — Phyrexian Vampire {4}{B}{B}{B}, 6/6:
//
//	"You may reveal this card from your opening hand. If you do, at the
//	 beginning of the first upkeep, each opponent loses 3 life, then you gain
//	 life equal to the life lost this way.
//	 Flying, lifelink"
//
// The reveal is the Chancellor seam (RevealFromOpeningHand, CR 103.6): the
// table is asked at the start of the game and a yes arms a delayed trigger
// for the first upkeep, whoever's it is. The gain is what was really lost.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8ce2a750-1137-4f8d-803c-9aac30db4193",
		Name:            "Chancellor of the Dross",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		OpeningHand: effects.RevealFromOpeningHand(game.RevealAtTheFirstUpkeep,
			"Chancellor of the Dross — each opponent loses 3 life, then you gain life equal to the life lost this way",
			chancellorOfTheDrossFirstUpkeep),
	})
}

var chancellorOfTheDrossFirstUpkeep = game.SimpleDelayedBody("chancellor-of-the-dross/first-upkeep-drain", func(g *game.Game, item *game.StackItem) error {
	return effects.B21DrainEachOpponentAndGainTheTotal(g, item, 3)
})
