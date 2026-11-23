package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emerge from the Cocoon — Sorcery {4}{W} (Duskmourn: House of Horror,
// #4466):
//
//	"Return target creature card from your graveyard to the battlefield.
//	 You gain 3 life."
//
// Zombify's return (reanimateSingleTarget, under the caster's control),
// then the life. A target gone by resolution leaves the spell with no
// legal target, so it does nothing at all, the life included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fed2127c-7174-4863-ad6a-1847c96bbd3f",
		Name:         "Emerge from the Cocoon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreatureInYourGraveyard(),
		Purpose:      effects.ForTargets(effects.Reanimates(0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			_, _ = effects.ReanimateSingleTarget(ctx, ctx.Controller())
			return effects.GainLife{Amount: 3}.Apply(ctx)
		},
	})
}
