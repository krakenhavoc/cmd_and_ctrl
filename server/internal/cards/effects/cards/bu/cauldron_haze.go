package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "56309bbe-86aa-4003-a531-0ef1e318ba15",
		Name:         "Cauldron Haze",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantEachLegalTargetUntilEOT(ctx, game.KeywordPersist, "Cauldron Haze — persist until end of turn")
		},
	})
}
