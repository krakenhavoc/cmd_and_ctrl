package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b78_helpers.go — shared pieces for the S82 GAP batch 78 cards
// (Ravnica Remastered enchantments and instants, #5948).

// gapB78HasPlusOneCounter is "a creature … with a +1/+1 counter on it"
// (Titanic Brawl): a card carrying at least one +1/+1 counter.
func GapB78HasPlusOneCounter() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.Counters[game.CounterPlusOne] > 0
	}
}
