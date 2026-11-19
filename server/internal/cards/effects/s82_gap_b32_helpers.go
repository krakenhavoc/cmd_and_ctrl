package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// gapB32EachPlayerDrawsACard is "each player draws a card" for a planeswalker
// ability's Effect: every player still in the game, you first (Jace Beleren).
func gapB32EachPlayerDrawsACard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range tablePlayers(ctx) {
		if err := (DrawCards{Player: id, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
