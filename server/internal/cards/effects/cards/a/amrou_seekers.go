package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Amrou Seekers — Creature — Kithkin Rebel {2}{W}, 2/2 (Time Spiral,
// #5596):
//
//	"This creature can't be blocked except by artifact creatures and/or
//	 white creatures."
//
// Manta Ray's evasion rule with a wider allowed set: a blocker that is
// an artifact or white (or both) may block it, and nothing else may.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "396efb56-7a54-4d91-b2ce-7f3c62c3238b",
		Name:         "Amrou Seekers",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnSelf(), effects.Or(effects.Artifact(), effects.OfColor("W")), "artifact creatures and/or white creatures"),
		},
	})
}
