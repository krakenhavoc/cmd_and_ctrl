package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b537_helpers.go — the Shadowmoor Wisps cycle (#5591): "Target
// creature becomes <colour> <rider> until end of turn. Draw a card."

// OldB537WispsResolve is a Wisps spell's resolution: the target's colours
// are replaced by `letter` (W, U, B, R or G) until end of turn (CR 105.3, a layer 5 effect),
// `rider` is applied to the same target (nil for none), then the caster
// draws a card. A target gone as the spell resolves fizzles the whole spell
// (CR 608.2b), draw included.
func OldB537WispsResolve(name, letter, word string, rider func(ctx *Context, target game.TargetRef) error) func(*game.StackItem, *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
			return nil
		}
		target := item.Targets[0]
		if err := UntilEndOfTurn(ctx, target.ID, nil, name+" — that creature becomes "+word,
			game.SetColorsMod(ctx.TextColors(letter)...)); err != nil {
			return err
		}
		if rider != nil {
			if err := rider(ctx, target); err != nil {
				return err
			}
		}
		return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
	}
}
