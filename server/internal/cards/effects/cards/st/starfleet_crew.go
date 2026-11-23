package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Starfleet Crew — {1}{W} Creature — Officer 2/3 (Star Trek Commander,
// #4974). The card has no rules text: it is a vanilla creature,
// registered with no abilities.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83b31215-ea9b-4ad9-a39e-3fe6cdb9b7de",
		Name:         "Starfleet Crew",
		Completeness: effects.CompletenessFull,
	})
}
