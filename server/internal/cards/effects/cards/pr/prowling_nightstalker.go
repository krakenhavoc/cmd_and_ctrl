package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prowling Nightstalker — Creature — Nightstalker {3}{B}, 2/2 (Portal Second
// Age, S82 older sets batch 733, #5625):
//
//	"This creature can't be blocked except by black creatures."
//
// CantBeBlockedExceptBy on itself (CR 509.1b); colour is read live.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6279078a-5c77-4a9e-8e9b-e99310363b6f",
		Name:         "Prowling Nightstalker",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnSelf(), effects.OfColor("B"), "black creatures"),
		},
	})
}
