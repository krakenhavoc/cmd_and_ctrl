package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// token_payoff.go — #3165: a creature token the bot makes is worth what
// its own creature-enter payoffs pay for it.
//
// ADR 0126's run 1 on the rebuilt mono-red-aggro deck (#3071) found two
// token makers used much less than the deck's others: Hordeling Outburst
// (three Goblins) was cast in 9 of the 37 games it was offered in, where
// Dragon Fodder and Krenko's Command, the same effect for two tokens,
// were cast in about two games in three; Krenko, Mob Boss's activation
// ({T}: create X Goblins) was taken in 8 of 30. In that deck every
// creature that enters is 2 damage to each opponent from Purphoros, and
// more from Impact Tremors and the other pingers, and nothing priced it.
//
// Two declarations, both read here:
//
//   - A creature-enter row declares `creature_enter_payoff`: what it pays
//     for each creature that enters under its controller's control
//     (Purphoros, Impact Tremors, Witty Roastmaster, the lifegain
//     creatures). It pays in landfall.go's units.
//   - A token maker declares `creature_tokens`, how many of its `tokens`
//     are creatures (a Treasure does not trigger Purphoros), or
//     `tokens_per_controlled`, a creature subtype whose count the bot
//     controls is the number it makes (Krenko, Mob Boss: X Goblins, X
//     the Goblins you control). The reader counts the second from the
//     view, since no printed number would be true.
//
// Each creature token is then worth the enter payoffs the bot's
// battlefield holds, and a count-at-resolution row is also worth
// TokenWeight per token, as a printed `tokens` is. A creature spell's own
// entry is not repriced: only the tokens a declared purpose makes.

// creatureEnterPayoff is what one creature entering under the bot's
// control pays it: the sum, over the triggered rows of the permanents it
// controls, of each declared creature-enter payoff. Zero when the Config
// does not price token payoffs.
func (st *state) creatureEnterPayoff(cfg Config) float64 {
	if !cfg.PriceTokenPayoffs || st.view == nil {
		return 0
	}
	var v float64
	cards := st.view.Battlefield.Cards
	for i := range cards {
		src := &cards[i]
		if src.Controller != st.me {
			continue
		}
		for _, r := range src.AbilityRows {
			if r.Kind != "triggered" || r.Purpose == nil || r.Purpose.CreatureEnterPayoff == nil {
				continue
			}
			l := protocol.LandfallPayoffView(*r.Purpose.CreatureEnterPayoff)
			v += st.landfallRowValue(cfg, &l)
		}
	}
	return v
}

// controlledOfType counts the creatures the bot controls with subtype
// `sub` (lowercase) on their type line.
func (st *state) controlledOfType(sub string) int {
	if st.view == nil {
		return 0
	}
	n := 0
	cards := st.view.Battlefield.Cards
	for i := range cards {
		c := &cards[i]
		if c.Controller == st.me && isCreature(c) && isType(c, sub) {
			n++
		}
	}
	return n
}

// tokenPayoffValue is what the creature tokens a declared purpose makes
// are worth beyond the TokenWeight a printed `tokens` already carries:
// the creature-enter payoffs per token, and TokenWeight per token for a
// count the reader works out itself.
func (p *Policy) tokenPayoffValue(st *state, ps purposeSet) float64 {
	if !p.cfg.PriceTokenPayoffs {
		return 0
	}
	n := ps.creatureTokens
	var v float64
	if ps.tokensPer != "" {
		k := st.controlledOfType(ps.tokensPer)
		n += k
		v += p.cfg.TokenWeight * float64(k)
	}
	if n <= 0 {
		return v
	}
	return v + float64(n)*st.creatureEnterPayoff(p.cfg)
}
