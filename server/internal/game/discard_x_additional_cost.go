package game

import "github.com/google/uuid"

// discard_x_additional_cost.go — "As an additional cost to cast this
// spell, discard X cards" (#6145): Channeled Force, Firestorm, Nahiri's
// Wrath, Abandon Hope, and with a card-type filter Scorched Earth ("X land
// cards") and Aether Tide ("X creature cards").
//
// None of these spells prints {X} in its mana cost. X is a value the
// caster chooses and announces as part of casting (CR 107.3a: "If a spell
// or activated ability has a mana cost, alternative cost, additional cost,
// and/or activation cost with an {X}, [-X], or X in it, and the value of X
// isn't defined by the text of that spell or ability, the controller of
// that spell or ability chooses and announces the value of X as part of
// casting the spell or activating the ability"), so it rides
// CastSpellParams.XValue like Toxic Deluge's "pay X life", and the stack
// item records it like any announced X: a target count bound to X ("Destroy
// X target lands") reads the same number.
//
// The discard itself is an ordinary cost discard (CR 601.2f: "Costs may
// include paying mana, tapping permanents, sacrificing permanents,
// discarding cards, and so on"): named on CastSpellParams.DiscardIDs,
// validated by validateAdditionalCostLocked against exactly X cards the
// clause accepts, paid while the spell is on the stack and recorded on
// PaidCost.Discarded, which is what "the discarded cards" reads (CR
// 608.2h). X = 0 is a legal announcement that discards nothing.

// DiscardCount is how many cards this cost discards when the cast
// announces `x`: the printed DiscardCards, or the announced X for a
// DiscardX clause (a negative X reads as 0). Nil-safe.
func (c *AdditionalCost) DiscardCount(x int) int {
	if c == nil {
		return 0
	}
	if c.DiscardX {
		return max(x, 0)
	}
	return c.DiscardCards
}

// DiscardXCeilingForEffect is the largest X a "discard X cards" cost lets
// `playerID` announce for the spell `castID`: the number of cards in their
// hand the clause accepts, the spell itself excluded (CR 601.2a has moved it
// to the stack). The view's X picker and the bot enumerator read it, so
// neither offers an X the validator refuses. 0 for every other cost.
//
// CALLER MUST ALREADY HOLD g.mu (read or write).
func (g *Game) DiscardXCeilingForEffect(playerID, castID uuid.UUID, cost *AdditionalCost) int {
	if cost == nil || !cost.DiscardX {
		return 0
	}
	return len(g.AdditionalDiscardOptionsForEffect(playerID, castID, cost))
}
