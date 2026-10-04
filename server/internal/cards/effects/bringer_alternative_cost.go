package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// bringer_alternative_cost.go — the Bringers' first line, shared by
// Bringer of the Black Dawn and Bringer of the Blue Dawn (and the rest
// of the cycle when they come):
//
//	"You may pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost."
//
// An alternative cost in CR 118.9's sense, declared as a bare
// game.AlternativeCost rather than through one of alternative_cost.go's
// keyword constructors: it is not a keyword, so there is no bundled
// rewrite (no cleared targets, no sacrifice trigger) for a constructor
// to own, and the creature resolves and enters exactly as one cast for
// {7}{B}{B} or {7}{U}{U} does. Baleful Mastery's shape, on a permanent
// spell.

// bringerAlternativeCost is the five-colour price. The key rides
// cast_spell as `alternative_cost` and lands on StackItem.AltCost.
func bringerAlternativeCost() game.AlternativeCost {
	return game.AlternativeCost{
		Key:      "bringer-wubrg",
		Label:    "Pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost",
		ManaCost: "{W}{U}{B}{R}{G}",
	}
}
