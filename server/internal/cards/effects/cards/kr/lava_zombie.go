package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lava Zombie — Creature — Zombie {1}{B}{R}, 4/3 (Planeshift, #5615):
//
//	"When this creature enters, return a black or red creature you
//	 control to its owner's hand.
//	 {2}: This creature gets +1/+0 until end of turn."
//
// The shared gating return (gating.go), chosen as it resolves, the
// Zombie itself included, and Fire Nation Cadets' generic-mana
// firebreathing: each activation is its own +1/+0, so they stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2232c38e-49eb-4823-a325-2342a952db72",
		Name:         "Lava Zombie",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.GatingReturn("Lava Zombie", "B", "R")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Lava Zombie — +1/+0 until end of turn"),
		}},
	})
}
