package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Glory Chaser — Creature — Goblin Warrior {R}, 1/1:
//
//	"Renown 1 (When this creature deals combat damage to a player, if
//	 it isn't renowned, put a +1/+1 counter on it and it becomes
//	 renowned.)
//	 As long as this creature is renowned, it has menace."
//
// #2049: renown is the engine's keyword trigger (game/renown.go); the
// menace is a layer-6 self-grant behind the Renowned gate.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cad1cd10-4810-495e-9305-0ae812596f6a",
		Name:            "Goblin Glory Chaser",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"renown 1"},
		Static:          []game.StaticAbility{RenownedKeywords("menace")},
	})
}
