package game

import "github.com/google/uuid"

// soulbond.go — soulbond's pairing (CR 702.95, Avacyn Restored; #5675):
// the one piece of state the keyword needs, and the rule that ends it.
//
//	CR 702.95b  A creature becomes "paired" with another as the result
//	            of a soulbond ability. Abilities may refer to a paired
//	            creature, the creature another creature is paired with,
//	            or whether a creature is paired. An "unpaired" creature
//	            is one that is not paired.
//	CR 702.95c  When the soulbond ability resolves, if either object
//	            that would be paired is no longer a creature, no longer
//	            on the battlefield, or no longer under the control of
//	            the player who controls the soulbond ability, neither
//	            object becomes paired.
//	CR 702.95d  A creature can be paired with only one other creature.
//	CR 702.95e  A paired creature becomes unpaired if any of the
//	            following occur: another player gains control of it or
//	            the creature it's paired with; it or the creature it's
//	            paired with stops being a creature; or it or the
//	            creature it's paired with leaves the battlefield.
//
// # The shape
//
// A pair is a link written on both creatures (Card.PairedWith, each
// naming the other). Nothing else is stored. The keyword's two
// triggered abilities (CR 702.95a) are catalog rows, effects.Soulbond,
// and they call PairForEffect as they resolve; a card's "as long as
// this is paired with another creature, each of those creatures has
// …" is an ordinary ADR 0093 layer-6 grant whose AppliesTo asks
// PairedPartner.
//
// # Lifetime (CR 702.95e)
//
// None of the three endings is a state-based action or an event the
// rules name: a pair simply stops. Two of them, the controller and the
// card types, are layered values, so the one place that sees all three
// settled is the end of the layer recompute, where
// sweepBrokenPairsLocked runs beside the untap holds' sweep. A pair it
// ends stays ended: a creature stolen until end of turn comes home
// unpaired, and an animated land that was paired is unpaired when it
// is a creature again.
//
// Phasing out is not leaving the battlefield and changes no controller
// (CR 702.26d), so a pair is kept while either creature is phased out.
// PairedPartner answers nothing in the meantime, because a phased-out
// permanent is treated as though it does not exist.
//
// Not copiable (CR 707.2): CopiableValuesOf never reads the link, so a
// copy of a paired creature is unpaired.

// KeywordSoulbond is CR 702.95's canonical token.
const KeywordSoulbond = "soulbond"

// PairedPartner is the creature `c` is paired with (CR 702.95b), or nil
// when c is unpaired, is not on the battlefield, or its partner is
// phased out. It never returns a half-link: the partner must name c
// back.
//
// Caller must hold g.mu. Read it after the layers are current.
func (g *Game) PairedPartner(c *Card) *Card {
	if c == nil || c.PairedWith == uuid.Nil {
		return nil
	}
	p := findBattlefieldCard(g, c.PairedWith)
	if p == nil || p.PairedWith != c.InstanceID {
		return nil
	}
	return p
}

// PairedWithForEffect is PairedPartner by ID: the partner's ID, or
// uuid.Nil for a permanent that is unpaired or not on the battlefield.
//
// Caller must hold g.mu in write mode.
func (g *Game) PairedWithForEffect(cardID uuid.UUID) uuid.UUID {
	g.RecomputeLayersIfStaleLocked()
	if p := g.PairedPartner(findBattlefieldCard(g, cardID)); p != nil {
		return p.InstanceID
	}
	return uuid.Nil
}

// CanPairForEffect reports whether a soulbond ability controlled by
// `controller` may pair the two named permanents now: they are two
// different creatures on the battlefield, both under that player's
// control (CR 702.95c) and both unpaired (CR 702.95d).
//
// Caller must hold g.mu in write mode.
func (g *Game) CanPairForEffect(controller, a, b uuid.UUID) bool {
	if a == b {
		return false
	}
	g.RecomputeLayersIfStaleLocked()
	return g.pairableLocked(controller, a) && g.pairableLocked(controller, b)
}

// pairableLocked is one half of CanPairForEffect.
func (g *Game) pairableLocked(controller, id uuid.UUID) bool {
	c := findBattlefieldCard(g, id)
	return c != nil && c.Controller == controller && c.IsCreature() && g.PairedPartner(c) == nil
}

// PairForEffect pairs two creatures (CR 702.95b) for a resolving
// soulbond ability controlled by `controller`. It reports whether it
// did; when CanPairForEffect says no, neither becomes paired
// (CR 702.95c) and that is ordinary play, not an error.
//
// Caller must hold g.mu in write mode.
func (g *Game) PairForEffect(controller, a, b uuid.UUID) bool {
	if !g.CanPairForEffect(controller, a, b) {
		return false
	}
	findBattlefieldCard(g, a).PairedWith = b
	findBattlefieldCard(g, b).PairedWith = a
	// The "as long as it is paired" grants come out of the layer pass.
	g.layerVersion.Add(1)
	return true
}

// UnpairedCreaturesForEffect lists the unpaired creatures `controller`
// controls, in battlefield order, leaving out `except`: the candidates
// of "pair this creature with another unpaired creature you control"
// (CR 702.95a).
//
// Caller must hold g.mu in write mode.
func (g *Game) UnpairedCreaturesForEffect(controller, except uuid.UUID) []uuid.UUID {
	if g.Battlefield == nil {
		return nil
	}
	g.RecomputeLayersIfStaleLocked()
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID != except && c.Controller == controller && c.IsCreature() && g.PairedPartner(c) == nil {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// sweepBrokenPairsLocked is CR 702.95e, run at the end of each layer
// recompute: every pair one of whose creatures changed controller this
// pass (`changed`, the deltas materialiseControlLocked found), is no
// longer a creature, or is gone, is unpaired on both sides. It reports
// whether it ended any, so the caller can ask for the pass that takes
// the pair's granted abilities away.
//
// A partner that is phased out is neither gone nor changed
// (CR 702.26d); the pair is judged again when both are phased in.
//
// Caller must hold g.mu in write mode.
func (g *Game) sweepBrokenPairsLocked(changed []controlChange) bool {
	if g.Battlefield == nil {
		return false
	}
	ended := false
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.PairedWith == uuid.Nil {
			continue
		}
		p := findBattlefieldCard(g, c.PairedWith)
		if p == nil {
			if g.phasedOutPartnerLocked(c) {
				continue
			}
			c.PairedWith = uuid.Nil
			ended = true
			continue
		}
		if p.PairedWith == c.InstanceID && c.IsCreature() && p.IsCreature() && c.Controller == p.Controller &&
			!controlChanged(changed, c.InstanceID) && !controlChanged(changed, p.InstanceID) {
			continue
		}
		if p.PairedWith == c.InstanceID {
			p.PairedWith = uuid.Nil
		}
		c.PairedWith = uuid.Nil
		ended = true
	}
	return ended
}

// phasedOutPartnerLocked reports whether c's partner is phased out and
// still names c.
func (g *Game) phasedOutPartnerLocked(c *Card) bool {
	if g.PhasedOut == nil {
		return false
	}
	for i := range g.PhasedOut.Cards {
		p := &g.PhasedOut.Cards[i]
		if p.InstanceID == c.PairedWith {
			return p.PairedWith == c.InstanceID
		}
	}
	return false
}

// controlChanged reports whether the materialise step moved control of
// the named permanent in this pass.
func controlChanged(changed []controlChange, id uuid.UUID) bool {
	for _, ch := range changed {
		if ch.card == id {
			return true
		}
	}
	return false
}
