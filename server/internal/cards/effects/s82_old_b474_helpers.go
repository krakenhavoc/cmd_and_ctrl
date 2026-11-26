package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b474_helpers.go — shared pieces of the Duel Decks: Elspeth vs.
// Tezzeret cards, S82 older-set batch 474 (#5578).

// S82OldB474TargetGainsFlyingUntilEndOfTurn is "target creature gains flying
// until end of turn" as a trigger body: the first chosen target still legal
// as the trigger resolves gets the keyword (CR 608.2b); a target that has
// left leaves the ability to do nothing. `name` labels the effect.
func S82OldB474TargetGainsFlyingUntilEndOfTurn(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: name + " — flying until end of turn"}.Apply(ctx)
	}
}

// S82OldB474CreaturesOfColor is the AppliesTo of "<Colour> creatures get
// …": every creature of the colour, whoever controls it. The colour is a
// word of the source's text (CR 612.2).
func S82OldB474CreaturesOfColor(color string) func(c *game.Card, g *game.Game, src *game.Card) bool {
	return func(c *game.Card, _ *game.Game, src *game.Card) bool {
		return c.IsCreature() && c.HasColor(src.TextColor(color))
	}
}
