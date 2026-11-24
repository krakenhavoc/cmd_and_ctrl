package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b013_helpers.go — shared pieces of the Secret Lair "Graveyard
// recursion" cards (S82 OLD batch 13).

// OldB013YourSliverCreatureCard is "each Sliver creature card in your
// graveyard": a creature card its controller owns with the Sliver type, the
// word read through the source's text (CR 612.2), changelings included.
func OldB013YourSliverCreatureCard(card *game.Card, _ *game.Game, source *game.Card) bool {
	if source == nil || card.Owner != source.Controller || !card.IsCreature() {
		return false
	}
	return card.HasSubtype(source.TextSubtype("Sliver"))
}
