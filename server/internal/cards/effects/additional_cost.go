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

// --- optional additional costs (ADR 0073, #664) -------------------
//
// An optional cost is the SAME struct with Optional set: what changes
// is that the caster chooses at CR 601.2b whether to pay it, and the
// choice is recorded so the resolution — or an entering permanent's
// own trigger — can read it back. One constructor per keyword, for
// the reason Overload and Evoke have one: the keyword carries the Key
// the engine reads, and a hand-rolled game.AdditionalCost{Optional:
// true} compiles and then never returns a bought-back card to hand.

// Kicker is CR 702.33's "Kicker [cost]" — "you may pay an additional
// [cost] as you cast this spell". Read back at resolution with
// ctx.WasKicked(), and from an entering permanent's own trigger with
// game.CardKickedTimes(*source).
//
//	OptionalCosts: []game.AdditionalCost{Kicker("{4}")},   // Burst Lightning
func Kicker(mana string) game.AdditionalCost {
	return game.AdditionalCost{
		Optional: true,
		Key:      game.KickerKey,
		ManaCost: mana,
		Label:    "Kicker " + mana,
	}
}

// KickerSacrifice is a NON-MANA kicker — "Kicker—Sacrifice a
// creature" (Gatekeeper of Malakir). The same Sacrifice component
// every mandatory sacrifice cost uses, so it is validated by the same
// validator and paid with the spell already on the stack: a Blood
// Artist drains before the Gatekeeper resolves.
func KickerSacrifice(label string, preds ...CardPredicate) game.AdditionalCost {
	return game.AdditionalCost{
		Optional:  true,
		Key:       game.KickerKey,
		Sacrifice: sacrificeSpec(label, preds...),
		Label:     "Kicker—Sacrifice " + label,
	}
}

// Multikicker is CR 702.33d's "Multikicker [cost]" — "you may pay an
// additional [cost] any number of times as you cast this spell".
// Read back with ctx.KickedTimes() and game.CardKickedTimes.
//
// `max` is the engine's cap on ONE announcement. It is not printed on
// any card — multikicker is unbounded in paper — but an announcement
// has to be finite, the client's stepper has to stop somewhere, and
// the bot's expansion has to terminate. A cap far above what any
// board can pay for is the honest place to put that; say so in the
// card's own comment.
//
//	OptionalCosts: []game.AdditionalCost{Multikicker("{G}", 20)},  // Wolfbriar Elemental
func Multikicker(mana string, max int) game.AdditionalCost {
	if max < 1 {
		max = 1
	}
	return game.AdditionalCost{
		Optional: true,
		Key:      game.MultikickerKey,
		ManaCost: mana,
		Repeat:   max,
		Label:    "Multikicker " + mana,
	}
}

// Buyback is CR 702.27's "Buyback [cost]" — "you may pay an
// additional [cost] as you cast this spell. If the buyback cost was
// paid, put this card into its owner's hand as it resolves."
//
// The return is the ENGINE's, not the card's: the resolution path
// reads the paid record and routes the spell to its owner's hand
// through the same stack-exit primitive flashback uses. A card file
// declares the cost and nothing else — and must NOT also return
// itself in OnResolve, which would move a card that is still on the
// stack.
//
//	OptionalCosts: []game.AdditionalCost{Buyback("{3}")},   // Capsize
func Buyback(mana string) game.AdditionalCost {
	return game.AdditionalCost{
		Optional: true,
		Key:      game.BuybackKey,
		ManaCost: mana,
		Label:    "Buyback " + mana,
	}
}

// BuybackSacrifice is a NON-MANA buyback — "Buyback—Sacrifice a land"
// (Constant Mists). The reason #664 exists at all: without it
// Constant Mists is a {1}{G} Fog.
func BuybackSacrifice(label string, preds ...CardPredicate) game.AdditionalCost {
	return game.AdditionalCost{
		Optional:  true,
		Key:       game.BuybackKey,
		Sacrifice: sacrificeSpec(label, preds...),
		Label:     "Buyback—Sacrifice " + label,
	}
}
