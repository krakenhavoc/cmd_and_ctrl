package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cached Defenses — Sorcery {2}{G} (Fate Reforged,
// #5552):
//
//	"Bolster 3. (Choose a creature with the least toughness among
//	 creatures you control and put three +1/+1 counters on it.)"
//
// Bolster (CR 701.39a) as the whole spell: the creature you control with the
// least toughness is chosen as it resolves, with no target, and with no
// creature nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e564f52-94fb-4c44-926f-0eebedc7d4c0",
		Name:         "Cached Defenses",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.Bolster{N: 3}),
	})
}
