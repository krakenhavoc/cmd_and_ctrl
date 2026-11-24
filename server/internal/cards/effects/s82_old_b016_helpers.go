package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b016_helpers.go — shared pieces of the Secret Lair "Instants and
// sorceries" cards (S82 OLD batch 16).

// OldB016BalanceSteps are Balance's three "same way" instructions, in printed
// order: lands, cards in hand, then creatures, each balanced to the player who
// has the fewest. It is Magus of the Balance's step (balanceStep) under
// Balance's own name, since the prompts name the card.
func OldB016BalanceSteps() []Applier {
	controlled := func(match func(game.Card) bool) func(*game.Game, uuid.UUID) int {
		return func(g *game.Game, player uuid.UUID) int { return CountControlled(g, player, match) }
	}
	return []Applier{
		balanceStep{label: "Balance — sacrifice lands down to the fewest any player controls",
			count: controlled(game.Card.IsLand), spec: SacrificeSpec("a land", Land())},
		balanceStep{label: "Balance — discard cards down to the fewest any player holds", discard: true,
			count: CardsInHand},
		balanceStep{label: "Balance — sacrifice creatures down to the fewest any player controls",
			count: controlled(game.Card.IsCreature), spec: SacrificeSpec("a creature", Creature())},
	}
}
