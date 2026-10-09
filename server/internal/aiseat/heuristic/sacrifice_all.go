package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// sacrifice_all.go — #2097: a spell whose additional cost sacrifices
// ALL of a kind of permanent the bot controls (Soulblast's "sacrifice
// all creatures you control").
//
// The policy cannot price the trade. sacrificeCost charges each
// sacrificed permanent its value, but what the spell BUYS is read from
// what it may target and what it costs (moves.go's opening note), and
// for Soulblast that is a six-mana "any target" spell whatever the
// board: the payoff the policy would credit does not grow with the
// creatures it gives up, and with none it would cast a six-mana spell
// that deals no damage. So the cast is declined outright, below
// passing, with a reason the decision log shows — a conservative gate
// rather than a price, until a declared purpose can say what the spell
// does per permanent it takes.

// sacrificeAllDeclined is what such a cast is worth: below passing, so
// the policy never takes it, as phyrexianLifeDeclined is.
const sacrificeAllDeclined = -1.0

// castSacrificesAll reports whether the card's additional cost takes
// every matching permanent the caster controls (the view's
// sacrifice_options.all). Nil-safe.
func castSacrificesAll(card *protocol.CardView) bool {
	if card == nil || card.AdditionalCost == nil || card.AdditionalCost.SacrificeOptions == nil {
		return false
	}
	return card.AdditionalCost.SacrificeOptions.All
}
