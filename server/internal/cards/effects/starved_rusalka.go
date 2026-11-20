package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starved Rusalka — Creature — Spirit {G}, 1/1 (Guildpact, #5599):
//
//	"{G}, Sacrifice a creature: You gain 1 life."
//
// High Market's outlet at {G}: any creature its controller controls,
// the Rusalka itself included, pays the cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0a57ac2e-276b-41ea-89f4-287f530114c6",
		Name:         "Starved Rusalka",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{G}, Sacrifice a creature: You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, LifeGain: 1},
			Cost:    Plus(ManaCost("{G}"), SacrificeACreature()),
			Effect:  Do(GainLife{Amount: 1}),
		}},
	})
}
