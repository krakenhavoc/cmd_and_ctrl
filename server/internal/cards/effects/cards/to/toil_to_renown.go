package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Toil to Renown — Sorcery {1}{G} (Shadowmoor, S82 OLD b529, #5591):
//
//	"You gain 1 life for each tapped artifact, creature, and land you
//	 control."
//
// The count is taken as the spell resolves; a permanent that is two of the
// three types is counted once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8323e46-c068-45e2-b6ba-7972b4af728a",
		Name:         "Toil to Renown",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GainLife{Player: ctx.Controller(), Amount: effects.OldB529TappedArtifactCreatureAndLandCount(ctx.Game, ctx.Controller())}.Apply(ctx)
		},
	})
}
