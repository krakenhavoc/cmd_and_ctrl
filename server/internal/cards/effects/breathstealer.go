package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Breathstealer — Creature — Nightstalker {2}{B}, 2/2 (Mirage, #5633):
//
//	"{B}: This creature gets +1/-1 until end of turn."
//
// The shared thisGetsUntilEndOfTurn body at +1/-1. The toughness loss is
// real: a third activation on an undamaged Breathstealer leaves it at 0
// toughness and it is put into its owner's graveyard as a state-based
// action.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17d18e2f-2d8d-449e-9356-a7351bf0f8c2",
		Name:         "Breathstealer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}: This creature gets +1/-1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: -1}},
			Cost:    ManaCost("{B}"),
			Effect:  thisGetsUntilEndOfTurn(1, -1, "Breathstealer — +1/-1 until end of turn"),
		}},
	})
}
