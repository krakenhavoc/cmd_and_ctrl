package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urborg Syphon-Mage — Creature — Human Spellshaper {2}{B}, 2/2 (Time
// Spiral, #5596):
//
//	"{2}{B}, {T}, Discard a card: Each other player loses 2 life. You gain
//	 life equal to the life lost this way."
//
// A Spellshaper (nemSpellshaperCost) whose spell is a drain of 2: every
// other player is an opponent at this table, and the gain is the life
// they really lost (b21DrainEachOpponentAndGainTheTotal).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d6608efe-b014-4b4d-a995-0be5047567c5",
		Name:         "Urborg Syphon-Mage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, {T}, Discard a card: Each other player loses 2 life. You gain life equal to the life lost this way.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    nemSpellshaperCost("{2}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b21DrainEachOpponentAndGainTheTotal(g, item, 2)
			},
		}},
	})
}
