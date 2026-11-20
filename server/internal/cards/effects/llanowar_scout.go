package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Llanowar Scout — Creature — Elf Scout {1}{G}, 1/3 (Dominaria, #5530):
//
//	"{T}: You may put a land card from your hand onto the battlefield."
//
// Sakura-Tribe Scout's ability: a land card from your hand onto the
// battlefield, if you choose one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3bdc9211-8eff-4b30-8d6f-d26f4087f77f",
		Name:         "Llanowar Scout",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: You may put a land card from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect:  Do(MayPutALandFromHand("Llanowar Scout")),
		}},
	})
}
