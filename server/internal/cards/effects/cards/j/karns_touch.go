package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Karn's Touch — Instant {U}{U} (Mercadian Masques, S82 OLD batch 697, #5619):
//
//	"Target noncreature artifact becomes an artifact creature with power and
//	 toughness each equal to its mana value until end of turn. (It retains its
//	 abilities.)"
//
// Sydri's animation (ArtifactBecomesCreatureOfItsManaValue): the mana value is
// read as the spell resolves (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5f9920d-7af5-4876-8823-6efbf2b4d9b9",
		Name:         "Karn's Touch",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target noncreature artifact", effects.And(effects.Artifact(), effects.Noncreature())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ArtifactBecomesCreatureOfItsManaValue(ctx, effects.DurationUntilEndOfTurn(ctx),
				"Karn's Touch — an artifact creature until end of turn")
		},
	})
}
