package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Happy Hogan, Dauntless Driver — {R} Legendary Creature — Human Pilot
// 2/1 (Marvel Super Heroes Commander, #4455). The card has no rules text:
// it is a vanilla creature, registered with no abilities.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "557ce9d4-8fe3-4ffc-8e56-b9b73d58d41b",
		Name:         "Happy Hogan, Dauntless Driver",
		Completeness: effects.CompletenessFull,
	})
}
