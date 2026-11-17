package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fallaji Excavation — Sorcery {3}{G}{G} (The Brothers' War, #5482):
//
//	"Create three tapped Powerstone tokens. You gain 3 life."
//
// Three tapped Powerstones made as one creation (CR 111.2), then the life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b13c91be-079a-46f1-8a99-b7dba99bb9c0",
		Name:         "Fallaji Excavation",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, LifeGain: 3},
		OnResolve: ResolveSteps(
			CreateTokenAdvanced{Spec: Token(PowerstoneToken()).EntersTapped(), N: 3},
			GainLife{Amount: 3},
		),
	})
}
