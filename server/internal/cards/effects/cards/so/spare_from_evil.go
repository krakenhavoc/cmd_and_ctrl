package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spare from Evil — Instant {1}{W} (Innistrad, S82 older sets batch 449,
// #5572):
//
//	"Creatures you control gain protection from non-Human creatures until
//	 end of turn."
//
// The token is game.ProtectionFromNonSubtypeCreatures("Human") (#6253); the
// set of creatures is locked at resolution (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "557cc9a7-c8cc-4823-b108-2dfe648c789d",
		Name:         "Spare from Evil",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{game.ProtectionFromNonSubtypeCreatures("Human")},
				Label:    "Spare from Evil — creatures you control gain protection from non-Human creatures",
			},
		),
	})
}
