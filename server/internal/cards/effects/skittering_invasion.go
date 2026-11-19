package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skittering Invasion — Kindred Sorcery — Eldrazi {7} (Rise of the Eldrazi,
// #5581):
//
//	"Create five 0/1 colorless Eldrazi Spawn creature tokens. They have
//	 "Sacrifice this token: Add {C}.""
//
// The tokens are the catalog's Eldrazi Spawn (tokens.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf890abd-b282-4dbb-ac86-9e7caec1c1b8",
		Name:         "Skittering Invasion",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 5, CreatureTokens: 5},
		OnResolve:    ResolveSteps(roeB1Spawns(5)),
	})
}
