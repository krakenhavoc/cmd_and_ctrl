package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// additional_cost.go — S21 sub-PR 5: constructors for Spec.
// AdditionalCost, the "As an additional cost to cast this spell, …"
// clause (CR 601.2f).
//
// Two shapes so far. Exile-from-graveyard still has no card in the
// current decklists, and a constructor with no caller is a guess about
// an API rather than an API.

// DiscardCost is "As an additional cost to cast this spell, discard
// N cards." The engine validates the caster's picks at announce and
// pays them with the spell already on the stack, so a discard
// payoff triggers above the spell and resolves first.
func DiscardCost(n int) *game.AdditionalCost {
	return &game.AdditionalCost{DiscardCards: n, Label: discardLabel(n)}
}

// discardLabel spells the clause the way the card prints it, since
// the client shows it verbatim above the picker.
func discardLabel(n int) string {
	switch n {
	case 1:
		return "Discard a card"
	case 2:
		return "Discard two cards"
	case 3:
		return "Discard three cards"
	}
	return "Discard " + strconv.Itoa(n) + " cards"
}

// SacrificeCost is "As an additional cost to cast this spell,
// sacrifice a creature" (Village Rites, Altar's Reap) — or any wider
// clause, via the predicates. Build the spec with the same
// sacrificeSpec helper the activated and mana abilities use, so
// "you may only sacrifice what you control" is enforced in one place.
//
// Paid with the spell already on the stack, so an aristocrats payoff
// watching the death triggers above the spell and drains first.
func SacrificeCost(label string, preds ...CardPredicate) *game.AdditionalCost {
	return &game.AdditionalCost{
		Sacrifice: sacrificeSpec(label, preds...),
		Label:     "Sacrifice " + label,
	}
}

// SacrificeNCost is "As an additional cost to cast this spell,
// sacrifice N <permanents>": SacrificeNCost(2, "two creatures",
// Creature()). SacrificeCost is the n = 1 case and stays the way to
// write it. The count rides on the clause exactly as SacrificeN's
// does (#747, ADR 0021 addendum), so the caster names exactly n
// permanents and they leave as one simultaneous exit while the spell
// is on the stack.
func SacrificeNCost(n int, label string, preds ...CardPredicate) *game.AdditionalCost {
	return &game.AdditionalCost{
		Sacrifice: sacrificeSpec(label, preds...).WithCount(n, n),
		Label:     "Sacrifice " + label,
	}
}

// PayXLifeCost is "As an additional cost to cast this spell, pay X
// life" — Toxic Deluge, and the third shape the clause takes.
//
// The X is announced with the cast and is the SAME number the spell's
// text reads back with ctx.X(): Toxic Deluge pays X life and gives
// every creature -X/-X, and a card that could announce those
// separately would be a different card. Paid with the spell already
// on the stack, like the other two components, so anything watching
// the life loss triggers above it.
func PayXLifeCost() *game.AdditionalCost {
	return &game.AdditionalCost{PayLifeX: true, Label: "Pay X life"}
}
