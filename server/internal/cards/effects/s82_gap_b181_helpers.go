package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b181_helpers.go — shared reads for GAP batch 181 (#5948,
// Starter Commander Decks).

// GapB181OtherFlyersYouControl counts the creatures `source`'s controller
// controls with flying other than `source` itself, after the layers.
func GapB181OtherFlyersYouControl(g *game.Game, source *game.Card) int {
	n := 0
	for _, id := range PermanentsControlledByMatching(g, source.Controller, And(Creature(), HasKeyword("flying"))) {
		if id != source.InstanceID {
			n++
		}
	}
	return n
}
