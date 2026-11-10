package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// std_b48_helpers.go — Standard set leftovers (#4452), batch 48.

// exploreTheModeTarget is one step of "that creature explores, then it
// explores again": the mode occurrence's target explores. Each explore is
// its own Instruction so the second waits for the first's graveyard
// question (CR 608.2c).
func exploreTheModeTarget(ctx *Context) error {
	t, ok := ModeTarget(ctx, ctx.Occurrence())
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return Explores{Explorer: t.ID}.Apply(ctx)
}

// returnTargetCardToTheBattlefield is "Return target <creature> card from
// your graveyard to the battlefield": the clause-0 target, if it is still
// legal (CR 608.2b), enters under its owner's control. Defossilize's first
// step and Too Evil to Stay Dead's whole body.
func returnTargetCardToTheBattlefield(_ *game.StackItem, ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
}
