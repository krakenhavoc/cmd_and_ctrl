package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b29_helpers.go — shared readers for Secrets of Strixhaven batch 29
// (#4456).

// socB29SpellWithXInItsManaCost is the CardPredicate "a spell with {X} in its
// mana cost" (Matterbending Mage): the cost printed on the card, parsed, so
// an {X} in an additional or alternative cost does not count (the same
// reading as socB9HasXInManaCost).
func socB29SpellWithXInItsManaCost(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	cost, err := game.ParseCost(c.ManaCost)
	return err == nil && cost.XSlots > 0
}
