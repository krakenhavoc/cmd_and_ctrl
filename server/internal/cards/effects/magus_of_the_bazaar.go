package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magus of the Bazaar — Creature — Human Wizard {1}{U}, 0/1 (Planar Chaos, #5595):
//
//	"{T}: Draw two cards, then discard three cards."
//
// Bazaar of Baghdad on a creature: drawThenDiscard draws two, then asks
// for three discards from the post-draw hand (CR 608.2c). The {T} keeps it
// off the turn it arrives (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e08d264d-5731-4e2a-ae6d-5a74ef648575",
		Name:         "Magus of the Bazaar",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Draw two cards, then discard three cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 2, Discards: 3},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return drawThenDiscard(g, item.Controller, item.SourceCardID, 2, 3, "")
			},
		}},
	})
}
