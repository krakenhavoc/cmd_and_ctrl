package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soulmender — Creature — Human Cleric {W}, 1/1 (Magic 2014, #5563):
//
//	"{T}: You gain 1 life."
//
// Braidwood Cup's ability on a 1/1. A tap ability of a creature, so summoning
// sickness applies (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "41b1d850-2fda-490b-a4f6-0d9bfb37103c",
		Name:         "Soulmender",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect:  gainOneLifeEffect,
		}},
	})
}
