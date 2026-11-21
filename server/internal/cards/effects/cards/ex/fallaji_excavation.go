package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fallaji Excavation — Sorcery {3}{G}{G} (The Brothers' War, #5482):
//
//	"Create three tapped Powerstone tokens. You gain 3 life."
//
// Three tapped Powerstones made as one creation (CR 111.2), then the life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b13c91be-079a-46f1-8a99-b7dba99bb9c0",
		Name:         "Fallaji Excavation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, LifeGain: 3},
		OnResolve: effects.ResolveSteps(
			effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 3},
			effects.GainLife{Amount: 3},
		),
	})
}
