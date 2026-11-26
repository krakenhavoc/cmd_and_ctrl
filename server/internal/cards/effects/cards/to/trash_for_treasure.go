package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trash for Treasure — Sorcery {2}{R} (Commander 2016, #5537):
//
//	"As an additional cost to cast this spell, sacrifice an artifact.
//	 Return target artifact card from your graveyard to the battlefield."
//
// Argivian Restoration's return with Forge Armor's additional sacrifice. The
// target is chosen before the cost is paid, so the sacrificed artifact is never
// a legal target of its own spell; a target that left the graveyard does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "e69279f6-50a1-4db5-83be-30a2ea7fc6a9",
		Name:           "Trash for Treasure",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("an artifact", effects.Artifact()),
		Targets:        effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.Artifact(), effects.YouOwn()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			_, _ = effects.ReanimateSingleTarget(ctx, ctx.Controller())
			return nil
		},
	})
}
