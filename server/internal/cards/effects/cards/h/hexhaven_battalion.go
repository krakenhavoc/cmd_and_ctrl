package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hexhaven Battalion — Sorcery {4}{W}{W} (Reality Fracture, tracker
// #2795):
//
//	"Create three 2/2 colorless Wizard Soldier creature tokens named
//	 Cadet. Empower Jace 2.
//	 Basic landcycling {2}"
//
// Three Cadets, then the keyword action (ADR 0139); basic landcycling is
// the ordinary typecycling activation from hand (#660).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "158823ed-6015-437a-82b0-8dadb1e9aab2",
		Name:         "Hexhaven Battalion",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3},
		// Empower Jace waits for a prompt the tokens' creation asks (a CR 616 order; CR 608.2c,
		// #4155).
		OnResolve: effects.ResolveSteps(
			effects.CreateToken{
				Template: effects.TokenCard("2/2 colorless Wizard Soldier named Cadet"),
				N:        3,
			},
			effects.EmpowerJace{N: 2},
		),
		Activated: []effects.ActivatedAbility{effects.BasicLandcycling("{2}")},
	})
}
