package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reduce to Dreams — Sorcery {3}{U}{U} (Betrayers of Kamigawa, #5602):
//
//	"Return all artifacts and enchantments to their owners' hands."
//
// Every artifact and enchantment on the battlefield, whoever controls it,
// returns in one simultaneous event (BounceAllMatching); a token that goes
// to a hand ceases to exist, and a commander's replacement prompt settles
// as it does for any mass bounce.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85bac6ad-b63a-4068-a813-ad685966e171",
		Name:         "Reduce to Dreams",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifactsAndEnchantments, How: game.SweepBounce}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BounceAllMatching{Match: effects.Or(effects.Artifact(), effects.Enchantment())}.Apply(ctx)
		},
	})
}
