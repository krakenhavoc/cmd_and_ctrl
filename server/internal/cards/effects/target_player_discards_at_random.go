package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// TargetPlayerDiscardsAtRandom is "target player discards a card at
// random" as an instruction (Flay): the first player among the item's
// targets that is still legal as it resolves (CR 608.2b) discards N cards
// at random (CR 701.9b).
type TargetPlayerDiscardsAtRandom struct {
	N int
}

func (s TargetPlayerDiscardsAtRandom) Apply(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return DiscardCards{Player: t.ID, N: s.N}.Apply(ctx)
		}
	}
	return nil
}
