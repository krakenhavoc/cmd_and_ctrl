package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// reveal_or_pay.go — the shared cost of the "reveal a
// <type> card from your hand or pay {N}" and "behold a <type> or pay
// {N}" cycles (ADR 0100 amendment 2026-10-07). Fifteen printed cards
// carry one of these two sentences, which differ only in the creature
// type, the mana and the "a" or "an", so the cost is built once here
// and every card names its three parameters:
//
//	AdditionalCost: RevealOrPay("an", "Elf", "{3}"), // Wren's Run Vanquisher
//	AdditionalCost: BeholdOrPay("a", "Merfolk", "{2}"), // Silvergill Mentor
//
// Branch 0 is the reveal (or behold) and branch 1 is the mana, in the
// order the cards print them. The keys are "reveal" / "behold" and
// "mana", for a card whose resolution reads ctx.PaidCostBranch (none of
// these does).

// RevealOrPay is "As an additional cost to cast this spell, reveal
// <article> <subtype> card from your hand or pay <mana>."
func RevealOrPay(article, subtype, mana string) *game.AdditionalCost {
	return EitherCost(
		RevealCardCost(article, subtype).Keyed("reveal"),
		ManaAdditionalCost(mana).Keyed("mana"),
	)
}

// BeholdOrPay is "As an additional cost to cast this spell, behold
// <article> <subtype> or pay <mana>." (To behold a Goblin, choose a
// Goblin you control or reveal a Goblin card from your hand.)
func BeholdOrPay(article, subtype, mana string) *game.AdditionalCost {
	return EitherCost(
		BeholdCost(article, subtype).Keyed("behold"),
		ManaAdditionalCost(mana).Keyed("mana"),
	)
}
