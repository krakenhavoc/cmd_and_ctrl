package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fetid Horror — Creature — Shade Horror {3}{B}, 1/2 (Mirage, #5633):
//
//	"{B}: This creature gets +1/+1 until end of turn."
//
// The shade pump, the shared thisGetsUntilEndOfTurn body: no tap in the
// cost, so any number of activations, summoning sick or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4863e1cf-3ef2-4c2d-a0ea-b73ca60c660d",
		Name:         "Fetid Horror",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    ManaCost("{B}"),
			Effect:  thisGetsUntilEndOfTurn(1, 1, "Fetid Horror — +1/+1 until end of turn"),
		}},
	})
}
