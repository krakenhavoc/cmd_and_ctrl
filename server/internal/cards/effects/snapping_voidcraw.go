package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snapping Voidcraw — Creature — Eldrazi Turtle {1}{G}{U}, 1/3 (Modern
// Horizons 3, #5467):
//
//	"Devoid (This card has no color.)
//	 {T}: Add {C}{C}.
//	 {3}{C}, {T}: Draw a card."
//
// Devoid is the engine's colour-removing keyword (CR 702.114a). The
// first ability is Hedron Archive's {C}{C} mana ability (CR 605.1a), and
// the second is Sea Gate Wreckage's {C}-costed draw without its
// condition: the {C} must be paid with colorless mana (CR 107.4c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e495033b-f727-41e3-b7fc-7f5fa8abdd87",
		Name:            "Snapping Voidcraw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}{C}, {T}: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    Plus(ManaCost("{3}{C}"), TapCost()),
			Effect:  b36DrawOne,
		}},
	})
}
