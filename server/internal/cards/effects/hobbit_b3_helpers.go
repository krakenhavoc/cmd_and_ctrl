package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shared predicates for The Hobbit cards, half B, batch 3 (#4454).

// hobB3TappedCreature is "a tapped creature" (Magnificent End).
func hobB3TappedCreature() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsCreature() && c.Tapped }
}

// hobB3AttackingNontokenCreature is "an attacking nontoken creature"
// (Uneasy Partings): declared as an attacker and not a token.
func hobB3AttackingNontokenCreature() CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		return AttackingCreature()(g, caster, c) && !c.IsToken()
	}
}

// hobB3PowerOrToughnessAtMost is "power or toughness N or less"
// (Stern Scolding), read live: on the stack a creature spell has the
// power and toughness it was printed with.
func hobB3PowerOrToughnessAtMost(n int) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.CurrentPower() <= n || c.CurrentToughness() <= n
	}
}
