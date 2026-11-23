package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Night Market Guard — Artifact Creature — Construct {3}, 3/1 (Aether Revolt,
// #5536):
//
//	"This creature can block an additional creature each combat."
//
// Foriysian Brigade's block capacity of two (CR 509.1a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a7a27fe-60f0-4a21-a1c1-5fef8a528391",
		Name:         "Night Market Guard",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.CanBlockAdditional(effects.SelfOnly, 1)},
	})
}
