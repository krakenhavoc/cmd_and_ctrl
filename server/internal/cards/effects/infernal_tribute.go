package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Infernal Tribute — Enchantment {B}{B}{B} (Weatherlight, #5629):
//
//	"{2}, Sacrifice a nontoken permanent: Draw a card."
//
// The sacrifice is part of the cost, so the activator names a nontoken
// permanent they control as they activate, Infernal Tribute itself
// included (CR 701.21a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a8bf79d2-29fd-4d2f-ac07-2e61d3886bc0",
		Name:         "Infernal Tribute",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice a nontoken permanent: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Draws: 1},
			Cost:    Plus(ManaCost("{2}"), SacrificeN(1, "a nontoken permanent", Not(IsTokenPredicate()))),
			Effect:  drawOneCardEffect,
		}},
	})
}
