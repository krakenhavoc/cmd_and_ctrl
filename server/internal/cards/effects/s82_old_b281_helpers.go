package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b281_helpers.go — shared shapes of S82 OLD batch 281 (Eternal
// Masters, #5542): the Serf token row and Void's mana-value predicate.

// The token row is added from this file's package-level initializer so the
// shared table's column alignment is not rewritten and it exists before any
// card file's init() asks for it. It matches the printed Serf token.
var _ = registerOldB281Tokens()

func registerOldB281Tokens() bool {
	for key, card := range map[string]game.Card{
		// Sengir Autocrat.
		"0/1 black Serf": {Name: "Serf", TypeLine: "Token Creature — Serf", Power: 0, Toughness: 1, Colors: []string{"B"}},
	} {
		if _, dup := TokenTable[key]; dup {
			panic("s82_old_b281_helpers: token key already in the table: " + key)
		}
		TokenTable[key] = card
	}
	return true
}

// ManaValueIs is a predicate for a card whose mana value is exactly n (Void).
func ManaValueIs(n int) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.ManaValue() == n }
}
