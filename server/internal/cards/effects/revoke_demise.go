package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Revoke Demise — Sorcery {4}{B} (Foundations Jumpstart, #5464):
//
//	"This spell costs {2} less to cast if it targets a creature card
//	 with mana value 3 or less.
//	 Return target creature card from your graveyard to the battlefield.
//	 You gain 2 life."
//
// Zombify's body with a life rider, and Price of Fame's self cost
// reduction (CostsLessIfItTargets) judged against the announced
// graveyard target as the cost is determined (CR 601.2f). A target gone
// by resolution leaves the spell with no legal target, so it does
// nothing, the life included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "83f0b573-58a0-457d-80e5-6e0b690e5f4e",
		Name:         "Revoke Demise",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
		Purpose:      ForTargets(Reanimates(0)),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a creature card with mana value 3 or less.",
				And(Creature(), ManaValueLE(3))),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			_, _ = reanimateSingleTarget(ctx, ctx.Controller())
			return GainLife{Amount: 2}.Apply(ctx)
		},
	})
}
