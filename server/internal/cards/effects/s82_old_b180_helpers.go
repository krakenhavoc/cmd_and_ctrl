package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// OldB180ShrinkEachBlocker is "each creature blocking it gets -1/-1 until end
// of turn" for a "becomes blocked" trigger on its own source (Plague Wight):
// every blocker of the source at resolution, as one group.
func OldB180ShrinkEachBlocker(label string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, id := range BlockersOf(g, item.SourceCardID) {
			if err := (BoostUntilEOT{Target: id, Power: -1, Toughness: -1, Label: label}).Apply(ctx.AsGroupMember()); err != nil {
				return err
			}
		}
		return nil
	}
}
