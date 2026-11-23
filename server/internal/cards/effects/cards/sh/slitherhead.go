package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slitherhead — Creature — Plant Zombie {B/G}, 1/1 (Return to Ravnica, #5566):
//
//	"Scavenge {0} ({0}, Exile this card from your graveyard: Put a
//	 number of +1/+1 counters equal to this card's power on target
//	 creature. Scavenge only as a sorcery.)"
//
// Scavenge is the shared graveyard ability (Deadbridge Goliath,
// scavenge.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b6966f6-052f-4d1d-a3bf-1d3cc91c71d4",
		Name:         "Slitherhead",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Scavenge("{0}")},
	})
}
