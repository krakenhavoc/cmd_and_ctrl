package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "fed2127c-7174-4863-ad6a-1847c96bbd3f",
		Name:         "Emerge from the Cocoon",
		Completeness: CompletenessFull,
		Targets:      targetCreatureInYourGraveyard(),
		Purpose:      ForTargets(Reanimates(0)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			_, _ = reanimateSingleTarget(ctx, ctx.Controller())
			return GainLife{Amount: 3}.Apply(ctx)
		},
	})
}
