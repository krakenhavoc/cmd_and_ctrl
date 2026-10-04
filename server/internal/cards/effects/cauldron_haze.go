package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cauldron Haze — Instant {1}{W/B}:
//
//	"Choose any number of target creatures. Each of those creatures
//	 gains persist until end of turn."
//
// Any number, zero included; each target still legal at resolution
// gains persist (#2075).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "56309bbe-86aa-4003-a531-0ef1e318ba15",
		Name:         "Cauldron Haze",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("any number of target creatures").WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return grantEachLegalTargetUntilEOT(ctx, game.KeywordPersist, "Cauldron Haze — persist until end of turn")
		},
	})
}
