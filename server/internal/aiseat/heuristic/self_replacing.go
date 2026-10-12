package heuristic

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// self_replacing.go is ADR 0126's amendment of 2026-10-12 (#2457): a
// card that can turn itself into a card is a cheaper discard.
//
// Marauding Mako's "Cycling {2}" is an activated ability that works from
// the hand: pay {2}, discard the Mako, draw a card. The owner's review
// of `do-not-windfall-away-the-last-land`: "if going to pay 2 to cycle
// mako might as well unexpected windfall the mako instead". A discard
// that a cost or an effect asks for does with the Mako what cycling it
// would have done, and spares the {2}.
//
// The bot reads it off the wire, never off the card's name: the hand
// card's `zone_abilities` row that discards the card itself
// (`discard_self`) and declares what it turns the card into. Cycling
// declares `draws: 1` and typecycling and transmute `tutors: 1`
// (cards/effects/cycling.go, transmute.go). A channel row, which also
// discards its card from the hand, turns it into an effect rather than
// a card and declares neither, so it is not one.
//
// The price, for such a nonland card:
//
//	replacement = Weights.Hand × (draws + TutorWeight × tutors)
//	spared      = SpellPerMana × the row's mana value
//	discard     = min(value, max(replacement, value − spared))
//
// where `value` is what the discard would otherwise cost (cardValue for
// a cost, handKeepValue for a discard the bot chooses). So the discard
// costs `spared` less, the mana the bot need not spend cycling it, and
// never less than the card cycling it would find, which the bot gives
// up with it. A big body with cycling keeps nearly all of its price; a
// 1/1 falls to about one card.
//
// A land is left alone. A land in hand is next turn's land drop, and
// cardValue, LastLandDiscard and DiscardLandFloor already say what that
// is worth; cycling a land gives the drop up too.

// selfReplacingDiscard returns what discarding the bot's own card c
// costs when its normal price is v (Config.SelfReplacingDiscards). With
// the switch off, or for a card that cannot turn itself into a card, it
// is v.
func (st *state) selfReplacingDiscard(cfg Config, c *protocol.CardView, v float64) float64 {
	if !cfg.SelfReplacingDiscards || c == nil || isLand(c) {
		return v
	}
	out := v
	for i := range c.ZoneAbilities {
		r := &c.ZoneAbilities[i]
		repl, spared, ok := st.selfReplacement(cfg, r)
		if !ok {
			continue
		}
		if p := min(v, max(repl, v-spared)); p < out {
			out = p
		}
	}
	return out
}

// selfReplacement reads one hand row: whether it turns its own card into
// a card, what that card is worth (replacement) and the mana it would
// spend doing so (spared), on the policy's value-per-mana proxy.
//
// A row qualifies when its cost is mana and discarding the card itself,
// nothing else, and its declared purpose draws or tutors. A row whose
// cost also asks for life, a sacrifice or other cards is a different
// trade and keeps the card's normal price.
func (st *state) selfReplacement(cfg Config, r *protocol.ActivatedAbilityView) (replacement, spared float64, ok bool) {
	if !r.DiscardSelf || r.Purpose == nil {
		return 0, 0, false
	}
	cards := float64(r.Purpose.Draws) + cfg.TutorWeight*float64(r.Purpose.Tutors)
	if cards <= 0 {
		return 0, 0, false
	}
	plain := *r
	plain.DiscardSelf = false
	if !rowCostsOnlyManaAndTaps(&plain) || plain.TapCost {
		return 0, 0, false
	}
	cost := r.ManaCost
	if r.ChargedManaCost != nil {
		cost = *r.ChargedManaCost
	}
	return st.w.Hand * cards, cfg.SpellPerMana * float64(manaValue(cost, 0)), true
}
