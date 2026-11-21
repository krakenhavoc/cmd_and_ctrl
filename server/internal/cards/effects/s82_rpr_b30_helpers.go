package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// counterTargetSpellAndExileIt is "Counter target spell. If that spell is
// countered this way, exile it instead of putting it into its owner's
// graveyard" as a spell's whole resolution (Assert Authority, Dissipate):
// the spell goes to exile through CounterTarget.Dest, and a spell that
// can't be countered stays on the stack and is not exiled either.
func CounterTargetSpellAndExileIt(_ *game.StackItem, ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return CounterTarget{StackID: t.ID, Dest: game.ZoneRef{Kind: game.ZoneExile}}.Apply(ctx)
}
