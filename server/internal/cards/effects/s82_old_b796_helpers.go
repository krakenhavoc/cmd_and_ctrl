package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB796ThisGetsMinusPerExtraBlocker is Jungle Wurm's "Whenever this creature
// becomes blocked, it gets -1/-1 until end of turn for each creature blocking it
// beyond the first": the blockers are counted as the trigger resolves, the
// mirror of rampage (CR 702.23).
func OldB796ThisGetsMinusPerExtraBlocker(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if !SourceIsStillThisPermanent(g, item) {
			return nil
		}
		extra := g.CreaturesBlockingForEffect(item.SourceCardID) - 1
		if extra <= 0 {
			return nil
		}
		return BoostUntilEOT{Target: item.SourceCardID, Power: -extra, Toughness: -extra, Label: label}.Apply(ctx)
	}
}
