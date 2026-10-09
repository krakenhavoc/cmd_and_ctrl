package heuristic

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// puts.go prices three things the 2026-10-08 review games (#2680,
// #2678) found the heuristic pricing as nothing, or as the opposite of
// what they are:
//
//   - A choose_cards prompt that puts the card it names from the bot's
//     hand onto the battlefield (Uro's land, Eureka Moment, Growth
//     Spiral). valueKeptInHand reads every choose_cards over the bot's
//     own hand as giving the named card up, so the bot declined every
//     one. The prompt now says where the card goes
//     (`choose_destination`), and the put is priced by what enters.
//   - An own_permanents pick (a karoo's "return a land you control to
//     its owner's hand", Lotus Field's sacrifice). It had no branch at
//     all, so every answer scored 0 and the enumerator's order chose.
//   - A declared extra land drop (`purpose.extra_land_drops`: Oracle of
//     Mul Daya, Exploration, Azusa), priced as nothing beyond the body.

// Choose-card destinations, as they appear on the wire
// (game.ChooseDestination).
const (
	chooseOntoBattlefield       = "battlefield"
	chooseOntoBattlefieldTapped = "battlefield_tapped"
)

// choiceOwnPermanents is game.PendingChoiceOwnPermanents on the wire.
const choiceOwnPermanents = "own_permanents"

// lateLandShare is the share of a mana source a land put onto the
// battlefield is worth once the bot has RampWantCap mana sources: the
// same 0.3 cardValue gives a late land in hand. Still positive, so a
// free land is still put.
const lateLandShare = 0.3

// valuePutOntoBattlefield scores one answer to a choose_cards prompt
// that puts the cards it names from the bot's own hand onto the
// battlefield. Every candidate is worth what keeping it in hand is worth
// (handKeepValue, the discard's scale), and a named one adds what it
// is worth on the battlefield over that: putGain. So the prompt is
// answered by the card that gains most, and declined only when nothing
// gains. The second return is false when the rule has no opinion: the
// Config is off, the prompt declares no battlefield destination, or a
// candidate is not in the bot's hand.
func (p *Policy) valuePutOntoBattlefield(st *state, ch *protocol.PendingChoiceView, named []string) (float64, bool) {
	if !p.cfg.PricePutsFromHand || ch == nil || len(ch.Options) == 0 || ch.FromPlayer != st.me || st.seat == nil {
		return 0, false
	}
	tapped := false
	switch ch.ChooseDestination {
	case chooseOntoBattlefield:
	case chooseOntoBattlefieldTapped:
		tapped = true
	default:
		return 0, false
	}
	hand := st.seatHand()
	held := make(map[string]*protocol.CardView, len(hand))
	for i := range hand {
		held[hand[i].InstanceID] = &hand[i]
	}
	var v float64
	for i := range ch.Options {
		c := held[ch.Options[i].InstanceID]
		if c == nil {
			return 0, false
		}
		v += st.handKeepValue(p.cfg, c)
		if slices.Contains(named, c.InstanceID) {
			v += p.putGain(st, c, tapped)
		}
	}
	return v, true
}

// putGain is what putting a card from hand onto the battlefield adds
// over keeping it. A land is a mana source the bot did not spend a land
// drop on: ManaSource (TappedManaSource when it enters tapped) and the
// ramp premium while the bot has fewer than RampWantCap mana sources,
// lateLandShare of that after; and the colours the hand needs
// (landColorFit, #2677) on top, so a dual is put before a basic. Any
// other permanent is its resolved value: what casting it would be worth,
// for no mana.
func (p *Policy) putGain(st *state, c *protocol.CardView, tapped bool) float64 {
	if !isLand(c) {
		// Resolving, not spent: its enters effect and an extra land
		// drop count as they would for a cast.
		return p.resolvedValueFor(st, c, 0, cardPurpose(c), false, false)
	}
	mana := st.w.ManaSource
	if tapped {
		mana = st.w.TappedManaSource
	}
	g := mana * lateLandShare
	if st.manaSources() < p.cfg.RampWantCap {
		g = mana + p.rampFor(st, c, 1)
	}
	return g + st.landColorFit(p.cfg, c)
}

// ownPermanentsValue scores one answer to an own_permanents pick: a
// choice among the bot's own permanents of the ones an effect bounces
// or sacrifices. Every printed fixed-count use spends what it names (a
// karoo's return, Lotus Field's two lands, annihilator), so the answer
// is worth minus what the named permanents are worth to keep
// (ownPermanentKeep), and the cheapest set goes.
//
// A pick whose count is the chooser's (Scapeshift's "any number", or
// Tragic Arrogance's own leg, where what is named is KEPT) has no sign
// this rule can read, so the second return is false and the
// enumerator's order still decides, as before.
func (p *Policy) ownPermanentsValue(st *state, ch *protocol.PendingChoiceView, named []string) (float64, bool) {
	if !p.cfg.PriceOwnPermanentPicks || ch == nil || ch.ChooseMin != ch.ChooseMax {
		return 0, false
	}
	var v float64
	for _, id := range named {
		v -= st.ownPermanentKeep(st.bf[id])
	}
	return v, true
}

// ownPermanentKeep is what keeping one of the bot's own permanents is
// worth when an effect asks which to give up: its fuel price
// (permanentValue), with a land priced by the mana it makes (a karoo
// is two lands' worth, which permanentValue's land arm does not count)
// and by its ability rows (a utility land is not a basic).
func (st *state) ownPermanentKeep(c *protocol.CardView) float64 {
	if c == nil {
		return 0
	}
	v := st.permanentValue(c)
	if isLand(c) {
		if n := repeatableMana(c); n > 1 {
			v *= float64(n)
		}
		v += st.w.rowUtility(c)
	}
	return v
}

// extraLandDropValue is what a declared extra land drop adds to a cast
// (#2678): one more land this turn for each drop the bot has a land in
// hand for that it could not otherwise play, at ManaSource and the ramp
// premium; and, for a permanent, ExtraLandDropRecurring per drop while
// the bot has fewer than RampWantCap mana sources, for the turns after.
// `self` is the card being cast, left out of the lands counted.
func (p *Policy) extraLandDropValue(st *state, n int, self *protocol.CardView) float64 {
	if !p.cfg.PriceExtraLandDrops || n <= 0 || st.seat == nil || self == nil {
		return 0
	}
	var v float64
	if st.myTurn {
		left := max(0, st.seat.LandDropsPerTurn-st.seat.LandsPlayedThisTurn)
		lands := 0
		for i := range st.seat.Hand.Cards {
			if c := &st.seat.Hand.Cards[i]; c.InstanceID != self.InstanceID && isLand(c) {
				lands++
			}
		}
		if extra := min(n, lands-left); extra > 0 {
			v += st.w.ManaSource*float64(extra) + p.rampFor(st, self, extra)
		}
	}
	if (isCreature(self) || isPermanentSpell(self)) && st.manaSources() < p.cfg.RampWantCap {
		v += p.cfg.ExtraLandDropRecurring * float64(n)
	}
	return v
}
