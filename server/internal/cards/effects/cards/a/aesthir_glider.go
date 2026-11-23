package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aesthir Glider — Artifact Creature — Bird Construct {3}, 2/1
// (Alliances, #5634):
//
//	"Flying
//	 This creature can't block."
//
// Flying is an engine keyword carried as printed data (CR 702.9b). "Can't
// block" is a restriction the permanent prints on itself, which the
// block declaration checks (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "99dcde0b-4f4d-4d7a-89aa-8a89c256e38f",
		Name:            "Aesthir Glider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})
}
