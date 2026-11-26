package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// CounterFirstTargetSpellThenDraw is "Counter target spell. Draw a card."
// (Contradict, Hindering Light): the first target is countered and its
// controller draws; with the target gone by resolution nothing happens and no
// card is drawn (CR 608.2b).
func CounterFirstTargetSpellThenDraw(item *game.StackItem, ctx *Context) error {
	if len(item.Targets) == 0 {
		return nil
	}
	if err := (CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
		return err
	}
	return DrawCards{N: 1}.Apply(ctx)
}
