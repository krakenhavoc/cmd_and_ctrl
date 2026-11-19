package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gapGraveyardCardsMatching counts the cards in `player`'s graveyard that
// satisfy `match`, as the spell resolves ("for each creature card in your
// graveyard", "for each instant and sorcery card in your graveyard"). A card
// in a graveyard has only its printed characteristics.
func gapGraveyardCardsMatching(g *game.Game, player uuid.UUID, match func(game.Card) bool) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if match(c) {
			n++
		}
	}
	return n
}
