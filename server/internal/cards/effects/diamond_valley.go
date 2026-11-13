package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Diamond Valley — Land (Arabian Nights, #5641):
//
//	"{T}, Sacrifice a creature: You gain life equal to the sacrificed
//	 creature's toughness."
//
// Miren, the Moaning Well's row with no mana in the cost. The toughness
// is the creature's as it last existed on the battlefield, pumps and
// counters included; a creature with 0 or less toughness gains nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "84cef34a-c3e1-4059-b4cd-c481938a53a5",
		Name:         "Diamond Valley",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice a creature: You gain life equal to the sacrificed creature's toughness.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(TapCost(), SacrificeACreature()),
			Effect:  gainLifeEqualToSacrificedToughness,
		}},
	})
}
