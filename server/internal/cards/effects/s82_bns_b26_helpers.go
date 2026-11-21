package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// targetPlayerMillsUntilLand is "target player reveals cards from the top
// of their library until they reveal a land card, then puts those cards
// into their graveyard" (Balustrade Spy, Undercity Informer): the
// ability's single player target mills.
func TargetPlayerMillsUntilLand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return b14MillUntilLand(ctx, t.ID)
		}
	}
	return nil
}
