package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// impulse.go is #2480: Professional Face-Breaker's "Sacrifice a
// Treasure: Exile the top card of your library. You may play that card
// this turn." Run 1 offered it in 1,070 windows and the bot took it in
// none, for two reasons, each behind a Config switch:
//
//   - THE CARD IT EXILES WAS NOT PRICED (Config.PriceImpulse). The row
//     declared only `answers: ["value"]`, so it was the flat
//     ActivateBase (0.50). It now declares `impulse: 1`, and an impulse
//     is priced as a card the bot can use this turn (impulseValue):
//     Weights.Hand per card in the bot's own main phase with an empty
//     stack, when it still has a land drop or two mana to play what it
//     finds, and nothing anywhere else, where the card is lost at the
//     end of the turn.
//   - A TREASURE GIVEN UP COST A LAND (Config.OneShotManaAsToken). A
//     sacrificed permanent costs its board value, and a Treasure's is
//     ManaSource (1.00), the price of a land or a rock that makes mana
//     every turn. Making one is worth TokenWeight (0.50, §6). A token
//     whose every mana ability sacrifices it makes its mana once, so
//     giving one up to pay a cost now costs what making one is worth
//     (sacrificedValue).
//
// Together, in the bot's own main phase: 1.20 for the card, less 0.50
// for the Treasure and the 0.30 every untapped creature's activation
// pays (moves.go), is +0.40, over PassThreshold. Either alone leaves it
// below zero.

// impulseMana is the untapped mana that counts as able to play an
// exiled card: most of the curated decks' spells cost two or more.
const impulseMana = 2

// impulseValue is what n cards exiled to be played this turn are worth
// to the bot now.
func (st *state) impulseValue(n int) float64 {
	if n <= 0 || !st.sorcerySpeed || st.seat == nil {
		return 0
	}
	// The pool and the repeatable sources only (manaAvailable): a
	// Treasure may be the one the cost sacrifices.
	landLeft := st.seat.LandDropsPerTurn > st.seat.LandsPlayedThisTurn
	if !landLeft && len(st.manaAvailable(false)) < impulseMana {
		return 0
	}
	return st.w.Hand * float64(n)
}

// oneShotManaToken reports whether c is a token that makes mana only by
// sacrificing itself: a Treasure, a Gold, a Powerstone's cousin with a
// sacrifice. A land, a creature and anything with a repeatable mana
// ability are not.
func oneShotManaToken(c *protocol.CardView) bool {
	if c == nil || !c.IsToken || len(c.ManaAbilities) == 0 || isLand(c) || isCreature(c) {
		return false
	}
	for i := range c.ManaAbilities {
		if !c.ManaAbilities[i].SacrificeCost {
			return false
		}
	}
	return true
}

// sacrificedValue is what permanent c is worth when the bot gives it up
// to pay a cost: its board value, or under OneShotManaAsToken
// TokenWeight for a one-shot mana token.
func (p *Policy) sacrificedValue(st *state, c *protocol.CardView) float64 {
	if p.cfg.OneShotManaAsToken && oneShotManaToken(c) {
		return p.cfg.TokenWeight
	}
	return st.permanentValue(c)
}
