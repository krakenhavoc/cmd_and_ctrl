package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// blast_if_its_color.go — the resolution of the Ice Age blasts,
// Pyroblast and Hydroblast (#5636):
//
//	"Choose one —
//	 • Counter target spell if it's <colour>.
//	 • Destroy target permanent if it's <colour>."
//
// "If it's <colour>" is a condition read as the spell resolves, not a
// restriction on the target, so the target clauses are unrestricted and
// a target of any other colour is left alone. The colour word is read
// through the spell's own text (ctx.TextColor), so a Sleight of Mind on
// the blast changes which colour it acts on.
//
// Append-only: add a builder, never change what one means.

// counterOrDestroyTheTargetIfItIs is the OnResolve body; `color` is the
// printed colour's symbol ("U" for Pyroblast, "R" for Hydroblast).
// Mode 0 counters a spell, mode 1 destroys a permanent.
func counterOrDestroyTheTargetIfItIs(color string) func(item *game.StackItem, ctx *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		if len(item.Targets) == 0 {
			return nil
		}
		target := item.Targets[0].ID
		card, ok := ctx.Game.LookupCardForEffect(target)
		if !ok || !card.HasColor(ctx.TextColor(color)) {
			return nil
		}
		switch {
		case ctx.HasMode(0):
			return CounterTarget{StackID: target}.Apply(ctx)
		case ctx.HasMode(1):
			if item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: target}.Apply(ctx)
		}
		return nil
	}
}
