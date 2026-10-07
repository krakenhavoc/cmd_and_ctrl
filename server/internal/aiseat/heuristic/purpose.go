package heuristic

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// purpose.go prices what a spell or an ability DOES, from the purpose
// the catalog declares for it (ADR 0126 §4, §6 and the discard half of
// §7).
//
// The wire never carried a spell's effect, so the heuristic priced
// every untargeted instant or sorcery by its mana value: Wrath of God
// and Divination were the same four-mana sorcery, and a wipe was cast
// onto the bot's own winning board as readily as onto a losing one. A
// purpose is the catalog's declared answer (`CardView.purpose`, on a
// mode, an alternative cost and an activated row too), and this file is
// the one place the policy reads it, so the cast price, the fuel price
// and the activation price agree about what one purpose is worth.
//
// Two halves:
//
//   - THE AMOUNTS (§6): cards drawn, searched out, milled on purpose,
//     discarded, lands put onto the battlefield, tokens made.
//
//     purposeValue = Hand × (draws + TutorWeight × tutors
//     + SelfMillWeight × selfMillTutor)
//     − DiscardWeight × discards
//     + ManaSource × lands + the ramp premium for those lands
//     + TokenWeight × tokens
//     + Weights.Energy × energy (ADR 0129 §7)
//
//     The lands get PR 3's ramp premium (rampPremium's deficit), so a
//     Rampant Growth is worth most while the bot cannot cast what it
//     holds and fades to a plain land when it can.
//
//   - THE SWEEP (§4): the change in ScoreEval with the permanents the
//     sweep would remove taken off each seat's board. That is what
//     stops the bot wiping its own winning board: the same number that
//     says a wipe into the table's leader is worth a lot says one into
//     the bot's own army is worth less than nothing.

// purposeSet is what one cast or activation declares it does, gathered
// from the slot the move names: the alternative cost it claims, the
// modes it chose, or the card itself. Amounts add; sweeps are kept
// apart, because two modes that each remove a class remove their union.
type purposeSet struct {
	draws, discards, lands, tutors, selfMill, tokens, energy int
	sweeps                                                   []protocol.SweepView
}

// add folds one declared purpose in.
func (ps *purposeSet) add(p *protocol.PurposeView) {
	if p == nil {
		return
	}
	ps.draws += p.Draws
	ps.discards += p.Discards
	ps.lands += p.Lands
	ps.tutors += p.Tutors
	ps.selfMill += p.SelfMillTutor
	ps.tokens += p.Tokens
	ps.energy += p.Energy
	if p.Sweep != nil {
		ps.sweeps = append(ps.sweeps, *p.Sweep)
	}
}

// hasAmounts reports whether any §6 amount is declared.
func (ps purposeSet) hasAmounts() bool {
	return ps.draws != 0 || ps.discards != 0 || ps.lands != 0 || ps.tutors != 0 ||
		ps.selfMill != 0 || ps.tokens != 0 || ps.energy != 0
}

// cardPurpose is the purpose the card itself declares: what the spell
// does, or a permanent's enters effect. It is the purpose a card being
// pitched to a cost is priced by, since no mode or alternative cost has
// been chosen for it.
func cardPurpose(c *protocol.CardView) purposeSet {
	var ps purposeSet
	if c != nil {
		ps.add(c.Purpose)
	}
	return ps
}

// castPurpose is the purpose of the cast a move names. ADR 0126 §4: an
// overloaded Cyclonic Rift carries its purpose on the alternative cost,
// and a modal wipe (Farewell, Austere Command) carries one per mode, so
// the move is priced by the cost or modes it names. A slot that
// declares nothing falls back to the card's own purpose.
func castPurpose(c *protocol.CardView, cp castParams) purposeSet {
	if c == nil {
		return purposeSet{}
	}
	if cp.AlternativeCost != "" {
		for i := range c.AlternativeCosts {
			if ac := &c.AlternativeCosts[i]; ac.Key == cp.AlternativeCost && ac.Purpose != nil {
				var ps purposeSet
				ps.add(ac.Purpose)
				return ps
			}
		}
	}
	if c.Modes != nil && len(cp.Modes) > 0 {
		var ps purposeSet
		found := false
		for _, m := range cp.Modes {
			if m >= 0 && m < len(c.Modes.Options) && c.Modes.Options[m].Purpose != nil {
				ps.add(c.Modes.Options[m].Purpose)
				found = true
			}
		}
		if found {
			return ps
		}
	}
	return cardPurpose(c)
}

// rowPurpose is the purpose an activated row of `src` declares, nil
// when it declares none. Looked up by the row's own index, which is
// what the activate move names.
func rowPurpose(src *protocol.CardView, index int) *protocol.PurposeView {
	if row := rowAt(src, index); row != nil {
		return row.Purpose
	}
	return nil
}

// rowAt is activatedRow (windows.go) for a source that may be nil.
func rowAt(src *protocol.CardView, index int) *protocol.ActivatedAbilityView {
	if src == nil {
		return nil
	}
	return activatedRow(src, index)
}

// purposePriced reports whether this Config prices the purpose at all:
// its amounts under PricePurposes, its sweeps under PriceSweeps. When it
// does not, the caller keeps the price it had before ADR 0126.
func (p *Policy) purposePriced(ps purposeSet) bool {
	return (p.cfg.PricePurposes && ps.hasAmounts()) || (p.cfg.PriceSweeps && len(ps.sweeps) > 0)
}

// purposeValue prices a declared purpose for the bot: the §6 amounts
// plus the §4 sweep. `x` is the announced X a sweep's amount may be,
// `self` the card being cast (left out of the ramp deficit, as
// rampPremium leaves it out), and clampSweep floors the sweep at zero
// for a card that is being SPENT rather than cast: a wipe that would
// hurt the bot today is still a card it might want tomorrow.
func (p *Policy) purposeValue(st *state, ps purposeSet, x int, self *protocol.CardView, clampSweep bool) float64 {
	var v float64
	if p.cfg.PricePurposes && ps.hasAmounts() {
		v += st.w.Hand * (float64(ps.draws) + p.cfg.TutorWeight*float64(ps.tutors) + p.cfg.SelfMillWeight*float64(ps.selfMill))
		v -= p.cfg.DiscardWeight * float64(ps.discards)
		// The discards may trigger the bot's own discard payoffs
		// (discard_payoff.go): Mary Read's loot with an Island in hand
		// makes a Treasure.
		v += st.resolutionDiscardPayoff(p.cfg, ps.discards, self)
		v += p.cfg.TokenWeight * float64(ps.tokens)
		// ADR 0129 §7: energy the effect gives, at the flat weight a
		// move spending it is charged.
		v += st.w.Energy * float64(ps.energy)
		if ps.lands > 0 {
			v += st.w.ManaSource * float64(ps.lands)
			v += p.rampFor(st, self, ps.lands)
		}
	}
	if p.cfg.PriceSweeps && len(ps.sweeps) > 0 {
		s := p.sweepValue(st, ps.sweeps, x)
		if clampSweep && s < 0 {
			s = 0
		}
		v += s
	}
	return v
}

// bounceShare is how much of a bounced permanent a sweep takes (ADR
// 0126 §4): half, because it comes back. A token does not come back
// (CR 111.7), so a bounce takes all of it.
const bounceShare = 0.5

// partialShare is how much of a matched permanent a `partial` sweep is
// presumed to take. Partial says the class is an upper bound — a
// nonwhite Doom Blade sweep, "without flying" — and the view does not
// say which permanents the condition spares, so the estimate splits
// the difference rather than pricing the bound as certain.
const partialShare = 0.5

// sweepValue is ADR 0126 §4's wipe price: ScoreEval with every
// permanent the sweeps would remove taken off its controller's board,
// less ScoreEval as it stands. The Hand a cast costs is charged by
// valueOfCast, as for every cast.
//
// ScoreEval already weighs the opposition by OpponentMean and
// OpponentMax, so a wipe that hits the table's leader is worth more,
// and one that takes the bot's own board while the opponents have
// little is negative. Several sweeps (two modes of Austere Command)
// remove the union of what each removes.
func (p *Policy) sweepValue(st *state, sweeps []protocol.SweepView, x int) float64 {
	if len(sweeps) == 0 || st.evals[st.me] == nil {
		return 0
	}
	cards := st.view.Battlefield.Cards
	share := make(map[string]float64, len(cards))
	for i := range cards {
		c := &cards[i]
		var s float64
		for j := range sweeps {
			if t := st.sweepShare(c, &sweeps[j], x); t > s {
				s = t
			}
		}
		if s > 0 {
			share[c.InstanceID] = s
		}
	}
	if len(share) == 0 {
		return 0
	}
	loss := map[string]float64{}
	for i := range cards {
		c := &cards[i]
		s := share[c.InstanceID]
		// An Aura goes with the permanent it enchants (CR 704.5m): the
		// +2/+2 on a dead creature, or the Pacifism that was answering
		// it. An Equipment stays behind, unattached.
		if h := st.attach.hostOf(c); h != nil && !isEquipment(c) && share[h.InstanceID] > s {
			s = share[h.InstanceID]
		}
		if s > 0 {
			loss[c.Controller] += s * st.permanentValue(c)
		}
	}
	after := make(map[string]*SeatEval, len(st.evals))
	for id, e := range st.evals {
		if l := loss[id]; l != 0 && !e.Eliminated {
			cp := *e
			cp.Board -= l
			cp.Strength -= l
			e = &cp
		}
		after[id] = e
	}
	return st.w.ScoreEval(after, st.me) - st.w.ScoreEval(st.evals, st.me)
}

// sweepShare is how much of permanent `c` one sweep removes: 1 for all
// of it, 0 for none, a fraction for a bounce or a partial sweep.
func (st *state) sweepShare(c *protocol.CardView, s *protocol.SweepView, x int) float64 {
	if s.OpponentsOnly && c.Controller == st.me {
		return 0
	}
	if e := st.evals[c.Controller]; e == nil || e.Eliminated {
		return 0
	}
	if !sweepMatches(c, s.Matches) {
		return 0
	}
	var share float64
	switch s.How {
	case "destroy":
		if hasKeyword(c, "indestructible") {
			return 0
		}
		share = 1
	case "exile", "sacrifice":
		share = 1
	case "bounce":
		share = bounceShare
		if c.IsToken {
			share = 1
		}
	case "damage", "minus":
		if !isCreature(c) {
			return 0
		}
		n := s.Amount
		if s.AmountIsX {
			n = x
		}
		if c.Toughness > n {
			return 0
		}
		if s.How == "damage" && hasKeyword(c, "indestructible") {
			return 0
		}
		share = 1
	default:
		return 0
	}
	if s.Partial {
		share *= partialShare
	}
	return share
}

// sweepMatches reports whether a permanent is in a sweep's class
// (game.SweepMatches). An unknown class matches nothing, so a value a
// later catalog adds is priced as no sweep rather than as every
// permanent.
func sweepMatches(c *protocol.CardView, m string) bool {
	switch m {
	case "creatures":
		return isCreature(c)
	case "nonland_permanents":
		return !isLand(c)
	case "artifacts":
		return isType(c, "artifact")
	case "enchantments":
		return isType(c, "enchantment")
	case "artifacts_and_enchantments":
		return isType(c, "artifact") || isType(c, "enchantment")
	case "all_permanents":
		return true
	case "creatures_mana_value_3_or_less":
		return isCreature(c) && manaValue(c.ManaCost, 0) <= 3
	case "creatures_mana_value_4_or_greater":
		return isCreature(c) && manaValue(c.ManaCost, 0) >= 4
	}
	return false
}

// discardCost is ADR 0126 §7's price for one card discarded to pay a
// spell's additional cost: what that card is worth to the bot
// (cardValue, the price the discard-choice branch already uses), not a
// flat card in hand. A late land costs little and an early one a lot.
//
// On top of it, the last land in hand while the bot is still short of
// LandsWanted costs LastLandDiscard more: it is next turn's land drop,
// and a hand with no land in it may miss one. That is what stops an
// Unexpected Windfall pitching the only land of an early hand.
func (p *Policy) discardCost(st *state, id string, castID string, discards []string) float64 {
	c := st.mine[id]
	if c == nil {
		return st.w.Hand
	}
	v := st.cardValue(p.cfg, c)
	if isLand(c) && p.cfg.LastLandDiscard != 0 && st.myMana < p.cfg.LandsWanted && st.seat != nil {
		kept := 0
		for i := range st.seat.Hand.Cards {
			h := &st.seat.Hand.Cards[i]
			if h.InstanceID == castID || !isLand(h) || contains(discards, h.InstanceID) {
				continue
			}
			kept++
		}
		if kept == 0 {
			v += p.cfg.LastLandDiscard
		}
	}
	return v
}

// discardsCost prices the cards an activated ability's discard cost
// names (#2016): discardCost per card, then each card's discard payoff
// back. With DiscardCostByCard off it is the flat Weights.Hand the cast
// branch charges.
func (p *Policy) discardsCost(st *state, sourceID string, ids []string) float64 {
	var v float64
	for _, id := range ids {
		if p.cfg.DiscardCostByCard {
			v += p.discardCost(st, id, sourceID, ids)
		} else {
			v += st.w.Hand
		}
		v -= st.discardPayoff(p.cfg, st.mine[id])
	}
	return v
}

// keywordCounters are the keyword counters (CR 122.1b) whose second
// copy buys nothing.
var keywordCounters = []string{"indestructible", "hexproof", "flying", "first strike", "deathtouch", "lifelink", "menace", "reach", "trample", "vigilance", "double strike"}

// redundantKeywordCounter reports whether the row at index puts a
// keyword counter on its own source that the source already has the
// keyword for. Read off the row's label ("Put an indestructible counter
// on Solphim") and the source's wire abilities: the wire declares no
// structured purpose for it, and CR 122.1b grants the keyword once.
func redundantKeywordCounter(src *protocol.CardView, index int) bool {
	row := rowAt(src, index)
	if row == nil {
		return false
	}
	label := strings.ToLower(row.Label)
	for _, kw := range keywordCounters {
		if (strings.Contains(label, "put a "+kw+" counter on") || strings.Contains(label, "put an "+kw+" counter on")) && hasKeyword(src, kw) {
			return true
		}
	}
	return false
}

func contains(ids []string, id string) bool {
	for _, s := range ids {
		if s == id {
			return true
		}
	}
	return false
}
