package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Terrus Wurm — Creature — Zombie Wurm {6}{B}, 5/5 (Return to Ravnica, #5566):
//
//	"Scavenge {6}{B} ({6}{B}, Exile this card from your graveyard: Put
//	 a number of +1/+1 counters equal to this card's power on target
//	 creature. Scavenge only as a sorcery.)"
//
// Scavenge is the shared graveyard ability (Deadbridge Goliath,
// scavenge.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0dfbee5a-0dc6-4f5b-abbd-9d8601de7107",
		Name:         "Terrus Wurm",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Scavenge("{6}{B}")},
	})
}
