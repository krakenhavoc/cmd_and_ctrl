package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glorifier of Dusk — Creature — Vampire Soldier {3}{W}{W}, 4/4 (Ravnica:
// Clue Edition, #5929):
//
//	"Pay 2 life: This creature gains flying until end of turn.
//	 Pay 2 life: This creature gains vigilance until end of turn."
//
// Two activated abilities with a life cost (CR 119.4: you can't pay more
// life than you have), each a keyword grant to itself until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6205c485-f46b-40af-8b73-03c668f9ea42",
		Name:         "Glorifier of Dusk",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.S82BnsB12GrantForLife("flying"),
			effects.S82BnsB12GrantForLife("vigilance"),
		},
	})
}
