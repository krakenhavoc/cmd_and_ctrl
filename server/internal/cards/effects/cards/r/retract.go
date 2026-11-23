package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Retract — Instant {U} (Darksteel, #5605):
//
//	"Return all artifacts you control to their owner's hand."
//
// Rebuild's mass bounce narrowed to its caster's artifacts. Each returns
// to its OWNER's hand, so a borrowed artifact goes home (CR 108.3). The
// set is read once, as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "31cefc9d-cc13-4b32-a1f8-edbcaf35aa3b",
		Name:         "Retract",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BounceAllMatching{Match: effects.And(effects.Artifact(), effects.YouControl())}.Apply(ctx)
		},
	})
}
