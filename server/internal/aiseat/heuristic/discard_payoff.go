package heuristic

import (
	"sort"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// discard_payoff.go — ADR 0126's amendment of 2026-10-06: a discard
// the bot's own permanents pay for is cheaper.
//
// Mary Read and Anne Bonny makes a tapped Treasure whenever her
// controller discards an Island, Pirate or Vehicle card, and Marauding
// Mako grows a +1/+1 counter for every card discarded. Before the
// amendment the bot priced a discard by what the card was worth in hand
// (cardValue) and nothing else, so with Mary Read out it pitched a
// Mountain as readily as an Island. A triggered row now declares
// `discard_payoff` (which discarded cards it pays on, and what it pays
// for each), and every place the policy prices one of its own cards
// being discarded subtracts what the bot's battlefield pays for it:
//
//   - a discard paid as a spell's additional cost (valueOfCast), or as
//     its alternative cost (ADR 0135 §2: Snag, Foil, retrace);
//   - a discard the bot chooses on resolution: a loot's, a rummage's,
//     "discard a card" (valueKeptInHand, choiceDiscardFromHand) and the
//     cleanup-step discard to hand size;
//   - the discards a declared purpose says a spell or an ability will
//     make (purposeValue), estimated from the cards the bot would pick.
//
// What one payoff is worth is priced in the units the policy already
// uses for the same thing: a token at TokenWeight (the Treasure a Big
// Score makes), a +1/+1 counter at Weights.Power + Weights.Toughness
// (what counterRemovalValue charges for losing one), and a point of
// damage to each live opponent at DamageToOpponent.

// discardPayoff is what discarding `c` pays the bot: the sum, over the
// triggered rows of the permanents it controls, of each discard payoff
// that matches the card. Zero when the Config does not price payoffs.
func (st *state) discardPayoff(cfg Config, c *protocol.CardView) float64 {
	if !cfg.PriceDiscardPayoffs || c == nil || st.view == nil {
		return 0
	}
	var v float64
	cards := st.view.Battlefield.Cards
	for i := range cards {
		src := &cards[i]
		if src.Controller != st.me || src.InstanceID == c.InstanceID {
			continue
		}
		for _, r := range src.AbilityRows {
			if r.Kind != "triggered" || r.Purpose == nil || r.Purpose.DiscardPayoff == nil {
				continue
			}
			d := r.Purpose.DiscardPayoff
			if !discardPayoffMatches(d, c) {
				continue
			}
			v += cfg.TokenWeight * float64(d.Tokens)
			v += (st.w.Power + st.w.Toughness) * float64(d.Counters)
			v += cfg.DamageToOpponent * float64(d.DamageEachOpponent*len(st.opps))
		}
	}
	return v
}

// discardPayoffMatches reports whether a discarded card is one the
// payoff pays on: any card, or a card with one of its types or subtypes
// on the type line, as a whole word ("Basic Land — Island",
// "Legendary Creature — Goblin Pirate").
func discardPayoffMatches(d *protocol.DiscardPayoffView, c *protocol.CardView) bool {
	if d.Any {
		return true
	}
	words := strings.FieldsFunc(strings.ToLower(c.TypeLine), func(r rune) bool {
		return r == ' ' || r == '—' || r == '-' || r == '/'
	})
	for _, t := range d.Types {
		for _, w := range words {
			if w == t {
				return true
			}
		}
	}
	return false
}

// resolutionDiscardPayoff estimates what the payoffs pay for the `n`
// cards a resolving purpose makes the bot discard (a loot's one,
// Faithless Looting's two). The bot will pitch the cards cheapest net
// of their payoff, which is how valueKeptInHand chooses, so the estimate
// is the payoff of those n cards in the hand it holds now. `self` is the
// card being cast, which will not be in hand to discard.
func (st *state) resolutionDiscardPayoff(cfg Config, n int, self *protocol.CardView) float64 {
	if n <= 0 || !cfg.PriceDiscardPayoffs {
		return 0
	}
	type cand struct{ net, pay float64 }
	var cands []cand
	hand := st.seatHand()
	for i := range hand {
		c := &hand[i]
		if self != nil && c.InstanceID == self.InstanceID {
			continue
		}
		pay := st.discardPayoff(cfg, c)
		cands = append(cands, cand{st.cardValue(cfg, c) - pay, pay})
	}
	sort.SliceStable(cands, func(i, j int) bool { return cands[i].net < cands[j].net })
	var v float64
	for i := 0; i < n && i < len(cands); i++ {
		v += cands[i].pay
	}
	return v
}

// altCostDiscards reports whether the alternative cost a cast claims
// discards the cards it is paid with (ADR 0135 §2): the offer's own
// `discards` flag on the card view. False for no claim or no card.
func altCostDiscards(c *protocol.CardView, key string) bool {
	if c == nil || key == "" {
		return false
	}
	for i := range c.AlternativeCosts {
		if ac := &c.AlternativeCosts[i]; ac.Key == key {
			return ac.Discards
		}
	}
	return false
}
