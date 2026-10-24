package heuristic

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// temporary_effect.go — #4302: a change to a target that ends with the
// turn, priced by what it does this turn and by whose permanent it is.
//
// Ian's report: with Heliod, Sun-Crowned out, a heuristic seat ignored the
// creature in its hand and its castable commander, and spent its mana on
// "{1}{W}: Another target creature gains lifelink until end of turn" aimed
// at an OPPONENT's creature, again and again until the mana ran out. Two
// causes, both in targetsValue's sign convention, "a target is hit":
//
//   - THE TARGET. With nothing on the wire to say what the row does to its
//     pick, an opposing creature was priced as removed (cardTargetValue:
//     its value × RemovalConfidence × LeaderBoost, about 2.5 for a 3/3)
//     and the bot's own as the small OwnPermanentTarget (0.40). Lifelink
//     on the opponent's creature was the best target on the list.
//   - THE RANKING. That removal price, plus the flat ActivateBase, beat a
//     two-drop's cast, and since the "removal" never removed anything the
//     same move was on offer, at the same price, in the next window. Only
//     the mana stopped it.
//
// The catalog now declares the direction on the target entry
// (TargetPurpose.UntilEndOfTurn: benefit or harm, with the printed Pump
// when there is one), and with Config.PriceTemporaryEffects the entry
// replaces targetsValue's guess for a permanent pick:
//
//   - a BENEFIT on an opponent's permanent, or a HARM on the bot's own, is
//     a gift to the wrong side: it costs the permanent's value, the price
//     a misplaced Aura pays (aura_host.go);
//   - a BENEFIT on the bot's own creature is worth the attack it adds this
//     turn (pumpAttackGain: more damage through, lifelink's life, an
//     evasion that slips the blockers, haste on a creature that could not
//     swing, LethalBonus when it is the kill), plus Config.CombatGrant
//     while that creature is attacking or blocking and damage is still to
//     come. Anywhere else it is worth nothing: it is gone at cleanup
//     (CR 514.2) having done nothing. One that adds nothing the creature
//     lacks (lifelink on a creature that has it, CR 702.15f) is below a
//     pass, and so is a no-amount grant ("can't be blocked this turn") the
//     same row already gave the same creature this turn, which the policy
//     remembers as free_repeat.go remembers a free activation;
//   - a HARM on an opponent's permanent is removal when its declared -N/-N
//     kills (CR 704.5f, 704.5g), the old price when it declares no amount,
//     and otherwise worth CombatGrant on a creature in combat and nothing
//     elsewhere.
//
// So outside combat such a move is worth what the attack gains, which a
// cast that develops the board beats, and in combat it clears the usual
// response bar only when it matters. With mana to spare it clears
// LeftoverThreshold there (spareCombatGrant): mana the bot's later casts
// this turn do not need is mana combat may spend.

// redundantGrant prices a change that gives its pick nothing it does not
// already have: below passing, like idleEquipMove.
const redundantGrant = -1.0

// temporaryEffectValue prices one permanent pick whose entry declares a
// change until end of turn. False when the entry declares none or the pick
// is not on the battlefield, which keeps the pick's ordinary price.
func (p *Policy) temporaryEffectValue(st *state, t targetRef, e *protocol.TargetPurposeView) (float64, bool) {
	if !p.cfg.PriceTemporaryEffects || e == nil || t.Kind == "player" {
		return 0, false
	}
	if e.UntilEndOfTurn != targetBenefit && e.UntilEndOfTurn != targetHarm {
		return 0, false
	}
	c := st.bf[t.ID]
	if c == nil {
		return 0, false
	}
	mine := c.Controller == st.me
	if mine == (e.UntilEndOfTurn == targetHarm) {
		// A benefit on an opponent's permanent, a harm on our own.
		return -st.hostWorth(p.cfg, c), true
	}
	if e.UntilEndOfTurn == targetBenefit {
		return p.temporaryBenefitValue(st, c, e.Pump), true
	}
	return p.temporaryHarmValue(st, c, e.Pump), true
}

// The wire names of game.TargetEffect.
const (
	targetBenefit = "benefit"
	targetHarm    = "harm"
)

// temporaryBenefitValue is what a change until end of turn is worth on the
// bot's own permanent c.
func (p *Policy) temporaryBenefitValue(st *state, c *protocol.CardView, pump *protocol.PumpView) float64 {
	if grantsNothingNew(c, pump) {
		return redundantGrant
	}
	gain, _ := p.pumpAttackGain(st, c, pump)
	if inCombatNow(st, c) {
		gain += p.cfg.CombatGrant
	}
	return gain
}

// temporaryHarmValue is what a change until end of turn is worth on an
// opponent's permanent c.
func (p *Policy) temporaryHarmValue(st *state, c *protocol.CardView, pump *protocol.PumpView) float64 {
	if pump == nil {
		// -X/-X, "can't block": no printed amount to test, so the old
		// reading, an opposing pick removed.
		return st.cardTargetValue(p.cfg, c.InstanceID)
	}
	if shrinkKills(c, pump) {
		return st.cardTargetValue(p.cfg, c.InstanceID)
	}
	if inCombatNow(st, c) {
		return p.cfg.CombatGrant
	}
	return 0
}

// grantsNothingNew reports whether a benefit adds nothing to c: no power
// or toughness, and every keyword it grants c already has (CR 702.15f:
// several instances of lifelink are redundant, and so for the others).
func grantsNothingNew(c *protocol.CardView, pump *protocol.PumpView) bool {
	if pump == nil || pump.Power != 0 || pump.Toughness != 0 || len(pump.Keywords) == 0 {
		return false
	}
	for _, k := range pump.Keywords {
		if !hasKeyword(c, k) {
			return false
		}
	}
	return true
}

// shrinkKills reports whether a harm's -N/-N kills creature c: toughness 0
// or less is put into the graveyard whatever else it has (CR 704.5f), and
// toughness at or below its marked damage destroys it unless it is
// indestructible (CR 704.5g, 702.12b).
func shrinkKills(c *protocol.CardView, pump *protocol.PumpView) bool {
	if !isCreature(c) || pump.Toughness >= 0 {
		return false
	}
	left := c.Toughness + pump.Toughness
	if left <= 0 {
		return true
	}
	return left <= c.DamageMarked && !hasKeyword(c, "indestructible")
}

// inCombatNow reports whether c is attacking or blocking with combat
// damage still to come.
func inCombatNow(st *state, c *protocol.CardView) bool {
	if c.AttackingTarget == "" && c.BlockingTarget == "" {
		return false
	}
	switch st.step {
	case "declare_attackers", "declare_blockers", "first_strike_damage":
		return true
	}
	return false
}

// temporaryPicks is each permanent pick of an activation or a cast whose
// declared entry is a change until end of turn, with that entry.
func (st *state) temporaryPicks(m legal.Move) ([]targetRef, []*protocol.TargetPurposeView) {
	var picks []targetRef
	var entries []*protocol.TargetPurposeView
	add := func(t targetRef, e *protocol.TargetPurposeView) {
		if t.Kind != "player" && e != nil && (e.UntilEndOfTurn == targetBenefit || e.UntilEndOfTurn == targetHarm) {
			picks = append(picks, t)
			entries = append(entries, e)
		}
	}
	switch m.Kind {
	case legal.KindActivate:
		cp := decode[activateParams](m.Params)
		src := st.bf[cp.SourceCardID]
		if src == nil {
			return nil, nil
		}
		for _, t := range cp.Targets {
			add(t, rowEntryFor(src, cp.AbilityIndex, t))
		}
	case legal.KindCast:
		cp := decode[castParams](m.Params)
		c := st.castSource(cp.InstanceID)
		if c == nil {
			return nil, nil
		}
		for _, t := range spellTargets(c, cp) {
			add(t, castEntryFor(c, cp, t))
		}
	}
	return picks, entries
}

// spareCombatGrant reports whether m is a change until end of turn on a
// creature in combat, paid with mana and taps only, that leaves the bot the
// mana for the biggest card it could still cast this turn. Such a move
// spends nothing the bot would otherwise keep, ADR 0126 §5's premise, and
// clears LeftoverThreshold. On another player's turn the bot's mana comes
// back in its untap step, so all of it is spare.
func (p *Policy) spareCombatGrant(st *state, m legal.Move) bool {
	if !p.cfg.PriceTemporaryEffects || !p.costsOnlyManaAndTaps(st, m) {
		return false
	}
	picks, _ := st.temporaryPicks(m)
	if len(picks) == 0 {
		return false
	}
	for _, t := range picks {
		c := st.bf[t.ID]
		if c == nil || !inCombatNow(st, c) {
			return false
		}
	}
	if !st.myTurn {
		return true
	}
	left := st.myMana - st.manaSpentBy(m)
	return left >= st.biggestCastWithin(st.myMana)
}

// manaSpentBy is the mana move m charges: the move's own declared cost
// when it carries one, else an activation's row cost as the engine
// charges it, X included.
func (st *state) manaSpentBy(m legal.Move) int {
	if m.Cost != nil && m.Cost.Mana != "" {
		return manaValue(m.Cost.Mana, 0)
	}
	if m.Kind != legal.KindActivate {
		return 0
	}
	cp := decode[activateParams](m.Params)
	src := st.bf[cp.SourceCardID]
	if src == nil {
		return 0
	}
	row := activatedRow(src, cp.AbilityIndex)
	if row == nil {
		return 0
	}
	cost := row.ManaCost
	if row.ChargedManaCost != nil {
		cost = *row.ChargedManaCost
	}
	return manaValue(cost, cp.XValue)
}

// biggestCastWithin is the largest mana value, its commander tax included,
// among the non-land cards in the bot's hand and command zone that `mana`
// could pay for: the most the bot's later casts this turn could want.
func (st *state) biggestCastWithin(mana int) int {
	if st.seat == nil {
		return 0
	}
	want := 0
	consider := func(c *protocol.CardView, tax int) {
		if isLand(c) {
			return
		}
		if mv := manaValue(c.ManaCost, 0) + tax; mv <= mana && mv > want {
			want = mv
		}
	}
	for i := range st.seat.Hand.Cards {
		consider(&st.seat.Hand.Cards[i], 0)
	}
	for i := range st.seat.Command.Cards {
		c := &st.seat.Command.Cards[i]
		consider(c, 2*st.seat.CommanderCasts[c.InstanceID])
	}
	return want
}

// noAmountGrant reports whether entry e is a benefit with no power or
// toughness to it, whose second copy on the same creature adds nothing
// the board can show ("can't be blocked this turn", protection from a
// colour chosen at resolution).
func noAmountGrant(e *protocol.TargetPurposeView) bool {
	return e != nil && e.UntilEndOfTurn == targetBenefit && (e.Pump == nil || (e.Pump.Power == 0 && e.Pump.Toughness == 0))
}

// grantKeys names each pick of m that noAmountGrant covers, by row and
// target: the memory's keys.
func (st *state) grantKeys(m legal.Move) []string {
	if m.Kind != legal.KindActivate {
		return nil
	}
	picks, entries := st.temporaryPicks(m)
	if len(picks) == 0 {
		return nil
	}
	row := freeActivationKey(decode[activateParams](m.Params))
	var keys []string
	for i, t := range picks {
		if noAmountGrant(entries[i]) {
			keys = append(keys, row+">"+t.ID)
		}
	}
	return keys
}

// repeatsTemporaryGrant reports whether m gives a creature a no-amount
// benefit the same row already gave it this turn.
func (p *Policy) repeatsTemporaryGrant(st *state, m legal.Move) bool {
	h := p.grantsThisTurn
	if !p.cfg.PriceTemporaryEffects || h == nil || h.seat != st.me || h.turn != st.turn {
		return false
	}
	keys := st.grantKeys(m)
	return len(keys) > 0 && !slices.ContainsFunc(keys, func(k string) bool { return !h.keys[k] })
}

// noteTemporaryGrant remembers the decision's no-amount grants. A new turn
// forgets the last one's.
func (p *Policy) noteTemporaryGrant(st *state, moves []legal.Move, d aiseat.Decision) {
	if h := p.grantsThisTurn; h != nil && (h.seat != st.me || h.turn != st.turn) {
		p.grantsThisTurn = nil
	}
	if !p.cfg.PriceTemporaryEffects || d.Index < 0 || d.Index >= len(moves) {
		return
	}
	keys := st.grantKeys(moves[d.Index])
	if len(keys) == 0 {
		return
	}
	if p.grantsThisTurn == nil {
		p.grantsThisTurn = &turnFreeActivations{seat: st.me, turn: st.turn, keys: map[string]bool{}}
	}
	for _, k := range keys {
		p.grantsThisTurn.keys[k] = true
	}
}
