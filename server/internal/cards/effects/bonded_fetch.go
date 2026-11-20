package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bonded Fetch — Creature — Homunculus {2}{U}, 0/2 (Future Sight, #5594):
//
//	"Defender, haste
//	 {T}: Draw a card, then discard a card."
//
// A looter with haste, so it can loot the turn it arrives. The loot is the
// shared draw-then-discard body; the discard is the controller's pick.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "79d34537-4633-4728-a1b3-5822119f0c84",
		Name:            "Bonded Fetch",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "haste"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return lootOne(g, item, 1)
			},
		}},
	})
}
