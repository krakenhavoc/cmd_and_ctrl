package heuristic

import (
	"math"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// card_choices.go prices the two card choices #2677 and #2691 found
// wrong in the 2026-10-08 review games: which land a search takes, and
// which card a discard from the bot's own hand gives up.
//
// Both are local to those decisions. cardValue, which the scry, the
// sacrifice, the fuel pricer and the cast-cost discard also read, is
// unchanged; a search adds landColorFit on top of it, and a discard
// reads handKeepValue in its place.

// --- #2677: a land search takes the colour the hand needs ---------------

// colorPips counts each single-coloured mana symbol in a cost: {G}{G}
// is two G. A hybrid or Phyrexian symbol ({G/U}, {2/G}, {G/P}) can be
// paid another way and is not a colour the card insists on, so it is
// left out; {C} is not a colour a land search can help with either.
func colorPips(cost string) map[string]int {
	out := map[string]int{}
	for {
		i := strings.IndexByte(cost, '{')
		if i < 0 {
			return out
		}
		j := strings.IndexByte(cost[i:], '}')
		if j < 0 {
			return out
		}
		sym := cost[i+1 : i+j]
		cost = cost[i+j+1:]
		switch sym {
		case "W", "U", "B", "R", "G":
			out[sym]++
		}
	}
}

// producedColors is the set of colours a card's repeatable mana
// abilities can add, read off ManaAbilities[].Produced the way
// repeatableMana reads amounts: "{G|U}" is G and U, "{W|U|B|R|G}" is all
// five. A land with no ability rows on the view makes nothing this can
// name.
func producedColors(c *protocol.CardView) map[string]bool {
	out := map[string]bool{}
	if c == nil {
		return out
	}
	for i := range c.ManaAbilities {
		ab := &c.ManaAbilities[i]
		if !ab.TapCost || ab.SacrificeCost || ab.ExileSelf || ab.AddsNoMana {
			continue
		}
		for _, r := range ab.Produced {
			switch r {
			case 'W', 'U', 'B', 'R', 'G':
				out[string(r)] = true
			}
		}
	}
	return out
}

// handColorNeed is the coloured pips across the bot's hand and its
// commander in the command zone.
func (st *state) handColorNeed() map[string]int {
	need := map[string]int{}
	if st.seat == nil {
		return need
	}
	for i := range st.seat.Hand.Cards {
		for k, n := range colorPips(st.seat.Hand.Cards[i].ManaCost) {
			need[k] += n
		}
	}
	for i := range st.seat.Command.Cards {
		for k, n := range colorPips(st.seat.Command.Cards[i].ManaCost) {
			need[k] += n
		}
	}
	return need
}

// colorSources counts, per colour, the bot's own battlefield permanents
// that can make it.
func (st *state) colorSources() map[string]int {
	have := map[string]int{}
	for _, c := range st.bf {
		if c.Controller != st.me {
			continue
		}
		for k := range producedColors(c) {
			have[k]++
		}
	}
	return have
}

// landColorUnneeded is what a colour the land makes is worth when
// nothing the bot holds asks for it: a little, so that among lands that
// meet the hand equally, the one making more colours (a dual over a
// basic) is taken.
const landColorUnneeded = 0.1

// landColorFit scores a land as a search target by the colours it makes
// (#2677). Each colour the bot's hand or commander asks for is worth
// 1/(1+sources), so the first source of a missing colour is worth the
// most and a fourth Island next to three is worth little; a colour
// nothing asks for is worth landColorUnneeded. Zero for a nonland or
// with LandColorNeed off, which leaves the search as it was: every land
// one flat cardValue, and the first offered taken.
func (st *state) landColorFit(cfg Config, c *protocol.CardView) float64 {
	if cfg.LandColorNeed == 0 || c == nil || !isLand(c) {
		return 0
	}
	need, have := st.handColorNeed(), st.colorSources()
	var fit float64
	for k := range producedColors(c) {
		if need[k] > 0 {
			fit += 1 / float64(1+have[k])
		} else {
			fit += landColorUnneeded
		}
	}
	return cfg.LandColorNeed * fit
}

// searchValue is what taking one card out of the bot's own library is
// worth: cardValue, plus a land's landColorFit.
func (st *state) searchValue(cfg Config, c *protocol.CardView) float64 {
	return st.cardValue(cfg, c) + st.landColorFit(cfg, c)
}

// --- #2691: a discard gives up what is furthest from castable -----------

// handKeepValue is what keeping a card in hand is worth when the bot
// chooses what to discard from it: the cleanup discard, a discard
// prompt, the cards a loot or a rummage names. With DiscardByDistance
// off it is cardValue.
//
// On, it differs from cardValue in three ways:
//
//   - A spell and a permanent are on one scale. An instant or sorcery
//     is worth DiscardSpellPerMana per mana (cardValue's SpellPerMana
//     is a cast-time proxy well under what a creature's body prices
//     per mana), so a cheap answer is no longer the first card out.
//   - Distance to castable replaces cardValue's flat ×0.6 for a card
//     more than one mana away: the card is discounted by
//     DistanceDiscount for each mana it is short, counting the lands in
//     hand as future drops and the colours it needs. A 7-drop on two
//     lands is five short; a Counterspell with one blue source is one.
//   - A land is worth what it brings the rest of the hand closer to
//     castable (landKeepValue), and never less than its cardValue, nor,
//     while the bot has fewer than RampWantCap mana sources, less than
//     DiscardLandFloor. The land is on the same scale as the spells it
//     is weighed against, so raising the spells does not tip the
//     discard onto the lands.
func (st *state) handKeepValue(cfg Config, c *protocol.CardView) float64 {
	if !cfg.DiscardByDistance || c == nil {
		return st.cardValue(cfg, c)
	}
	sources := st.manaSources()
	if isLand(c) {
		v := max(st.cardValue(cfg, c), st.landKeepValue(cfg, c, sources))
		if sources < cfg.RampWantCap {
			// A Commander deck wants its seventh mana: below it, a land
			// is a future land drop whatever the hand holds now.
			v = max(v, cfg.DiscardLandFloor)
		}
		return v
	}
	return st.spellKeepValue(cfg, c, sources, "")
}

// spellKeepValue is a nonland card's keep value, with the land
// `without` (if any) taken out of the hand.
func (st *state) spellKeepValue(cfg Config, c *protocol.CardView, sources int, without string) float64 {
	var v float64
	switch {
	case isCreature(c) || isPermanentSpell(c):
		v = st.w.permanentValue(c)
	default:
		v = cfg.DiscardSpellPerMana * float64(manaValue(c.ManaCost, 0))
	}
	if short := st.manaShort(c, sources, without); short > 0 {
		v *= math.Pow(cfg.DistanceDiscount, float64(short))
	}
	return v
}

// landKeepValue is what a land in hand adds to the rest of the hand: for
// each nonland card, its keep value with the land minus its keep value
// without it. Early, with the hand short of mana, that is most of a
// card; with every card castable, or another land in hand covering the
// same gap, it is nothing.
func (st *state) landKeepValue(cfg Config, land *protocol.CardView, sources int) float64 {
	var v float64
	hand := st.seatHand()
	for i := range hand {
		h := &hand[i]
		if isLand(h) {
			continue
		}
		v += st.spellKeepValue(cfg, h, sources, "") - st.spellKeepValue(cfg, h, sources, land.InstanceID)
	}
	return v
}

// manaSources is the mana the bot's own sources make, tapped or not:
// rampFor's `sources`.
func (st *state) manaSources() int {
	n := 0
	for _, c := range st.bf {
		if c.Controller == st.me {
			n += repeatableMana(c)
		}
	}
	return n
}

// manaShort is how many mana the bot is short of casting c with what it
// has and the lands it holds, leaving out the land `without` (#2691): the
// larger of the generic gap (mana value less sources and lands in hand)
// and the colour gap (each colour's pips less the sources and lands in
// hand that make it).
func (st *state) manaShort(c *protocol.CardView, sources int, without string) int {
	landsInHand := 0
	handColors := map[string]int{}
	hand := st.seatHand()
	for i := range hand {
		h := &hand[i]
		if !isLand(h) || h.InstanceID == c.InstanceID || h.InstanceID == without {
			continue
		}
		landsInHand++
		for k := range producedColors(h) {
			handColors[k]++
		}
	}
	short := manaValue(c.ManaCost, 0) - sources - landsInHand
	have := st.colorSources()
	colorShort := 0
	for k, n := range colorPips(c.ManaCost) {
		if gap := n - have[k] - handColors[k]; gap > 0 {
			colorShort += gap
		}
	}
	return max(short, colorShort, 0)
}
