package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB350TargetPlayerDiscards is "target player discards N cards" as one
// resolution step: the targeted player chooses what they discard (CR 701.9b)
// as a real prompt, and a spell whose target has become illegal does nothing
// (CR 608.2b). Rakshasa's Secret's first sentence.
func OldB350TargetPlayerDiscards(n int) func(ctx *Context) error {
	return func(ctx *Context) error {
		player := TargetedPlayer(ctx)
		if player == uuid.Nil {
			return nil
		}
		ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: player, Source: ctx.Item.SourceCardID, N: n})
		return nil
	}
}

// OldB350TargetGetsMinusXUntilEOT is "target creature gets -X/-X until end
// of turn", X being the counters removed as the cost was paid (Retribution of
// the Ancients). Nothing happens for X of 0 or when the target is gone.
func OldB350TargetGetsMinusXUntilEOT(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := ctx.CountersRemoved()
	if x <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return BoostUntilEOT{Target: t.ID, Power: -x, Toughness: -x}.Apply(ctx)
		}
	}
	return nil
}
