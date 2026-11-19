package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gnawing Zombie — Creature — Zombie {1}{B}, 1/3 (Foundations Jumpstart):
//
//	"{1}{B}, Sacrifice a creature: Target player loses 1 life and you
//	 gain 1 life."
//
// The sacrifice is Viscera Seer's cost (the Zombie may sacrifice
// itself); the body is drainTargetOpponentOne, which drains the first
// still-legal player target (CR 608.2b). No simplification.
func init() {
	Register(Spec{
		OracleID:     "7979c1e6-6bc5-4b8c-82b0-2770e67ba898",
		Name:         "Gnawing Zombie",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice a creature: Target player loses 1 life and you gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, LifeGain: 1},
			Cost:    Plus(ManaCost("{1}{B}"), SacrificeACreature()),
			Targets: TargetPlayer("target player"),
			Effect:  drainTargetOpponentOne,
		}},
	})
}
