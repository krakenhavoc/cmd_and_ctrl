package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// fin_b1_helpers.go — shared helpers for Final Fantasy, batch 1 (#4462).

// finB1CondEightInYourGraveyard is the intervening "if eight or more
// cards are in your graveyard" (Sidequest: Card Collection), re-checked
// as the ability resolves (CR 603.4). It reads the item's controller's
// graveyard (game.FootprintCondGraveyard).
type finB1CondEightInYourGraveyard struct{}

func (finB1CondEightInYourGraveyard) Holds(ctx *Context) bool {
	return GraveyardAtLeast(8, nil)(ctx.Game, ctx.Controller(), ctx.Source())
}

func (finB1CondEightInYourGraveyard) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondGraveyard, 0), true
}

// finB1TapAndStun is "tap target <permanent>. Put a stun counter on it":
// the announced target, if still legal (Ice Flan).
func finB1TapAndStun(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return AddCounter{Target: id, Kind: game.CounterStun, N: 1}.Apply(ctx)
}
