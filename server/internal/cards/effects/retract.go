package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "31cefc9d-cc13-4b32-a1f8-edbcaf35aa3b",
		Name:         "Retract",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BounceAllMatching{Match: And(Artifact(), YouControl())}.Apply(ctx)
		},
	})
}
