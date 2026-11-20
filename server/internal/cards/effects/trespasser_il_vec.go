package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trespasser il-Vec — Creature — Human Rogue {2}{B}, 3/1 (Time Spiral,
// #5596):
//
//	"Discard a card: This creature gains shadow until end of turn."
//
// Thalakos Drifters' row: the discard is the cost (CR 602.2b), and the
// shadow lasts until the cleanup step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "90980841-4994-453c-a076-6cdf8d85c809",
		Name:         "Trespasser il-Vec",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Discard a card: This creature gains shadow until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    DiscardACard(),
			Effect:  thisGainsKeywordUntilEndOfTurn("shadow", "Trespasser il-Vec — shadow"),
		}},
	})
}
