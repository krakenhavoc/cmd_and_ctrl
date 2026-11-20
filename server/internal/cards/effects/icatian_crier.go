package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Icatian Crier — Creature — Human Spellshaper {2}{W}, 1/1 (Time Spiral,
// #5596):
//
//	"{1}{W}, {T}, Discard a card: Create two 1/1 white Citizen creature
//	 tokens."
//
// A Spellshaper (nemSpellshaperCost) whose spell is half an Icatian Town.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "668312dd-7b6e-46bf-9c6a-bfb03dd75b8f",
		Name:         "Icatian Crier",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}, {T}, Discard a card: Create two 1/1 white Citizen creature tokens.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    nemSpellshaperCost("{1}{W}"),
			Effect:  Do(CreateToken{Template: TokenCard("1/1 white Citizen"), N: 2}),
		}},
	})
}
