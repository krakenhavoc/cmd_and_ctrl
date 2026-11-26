package game

import "github.com/google/uuid"

// licid.go — the licids (#5781, ADR 0154 amendment 2026-11-25):
//
//	Leeching Licid: "{B}, {T}: This creature loses this ability and
//	becomes an Aura enchantment with enchant creature. Attach it to
//	target creature. You may pay {B} to end this effect."
//
// # Becoming an Aura
//
// It is the Aura form a permanent already on the battlefield takes when it
// "becomes an Aura with enchant creature" (BecomeAuraForEffect, Last
// Voyage of the _____): Card.ReturnedAura, which the card's own statics
// read (effects.LicidForm: the layer-4 type set, CR 205.1a, and the layer-6
// loss of the one activated row, CR 613.1f), and CardDef.ReturnedAura as
// the enchant clause CR 704.5m holds it to. BecomeLicidAuraForEffect sets
// the form, records who may end it and for what, and attaches the licid to
// the target (CR 701.3a). It is an Aura that was never an Aura spell, so it
// targets nothing as it attaches; an Aura whose host leaves the
// battlefield, or that cannot legally enchant it, is put into its owner's
// graveyard by the state-based action (CR 704.5m), as the licid rulings
// say.
//
// # Ending the effect (CR 116.2c)
//
// "You may pay {B} to end this effect" is CR 116.2c's special action:
//
//	116.2c Some effects allow a player to take an action at a later time,
//	usually to end a continuous effect or to stop a delayed triggered
//	ability from triggering. Doing so is a special action. A player can
//	take such an action any time they have priority, unless that effect
//	specifies another timing restriction, for as long as the effect
//	allows it.
//
// The kind is end_effect, an offer DERIVED from the permanent (like
// unlock): a permanent in the Aura form with AuraFormEndCost set offers it
// to AuraFormEndBy, the player who activated the ability that made the
// effect (CR 109.5: for an activated ability, "you" is "the player who
// activated the ability"), whoever controls the licid now. Any time that player
// has priority, on any turn, with anything on the stack; it is not a cast
// or an activation, so split second does not stop it (CR 702.61b). Taking
// it ends the effect: the Aura form and the offer are cleared, so the
// licid is a creature again with its ability back, and it becomes
// unattached (CR 704.5p: "If a battle or creature is attached to an object
// or player, it becomes unattached and remains on the battlefield"; the
// action does it at once rather than leaving a creature attached until the
// next check). It stays on the battlefield and has been under its
// controller's control all along.

// SpecialActionEndEffect is CR 116.2c: pay to end the effect that made a
// licid an Aura. Its card_id is the licid.
const SpecialActionEndEffect SpecialActionKind = "end_effect"

// EndEffectOffer is the end_effect offer of a permanent in a licid's Aura
// form, or nil. Derived from the card, like UnlockOffers: nothing in the
// catalog declares it.
func EndEffectOffer(c Card) *SpecialAction {
	if !c.ReturnedAura || c.AuraFormEndBy == uuid.Nil || c.AuraFormEndCost == "" {
		return nil
	}
	return &SpecialAction{
		Kind:  SpecialActionEndEffect,
		Cost:  c.AuraFormEndCost,
		Label: "Pay " + c.AuraFormEndCost + " to end the effect",
	}
}

// BecomeLicidAuraForEffect is the licid's ability resolving: the licid
// takes the Aura form (it loses its ability and becomes an Aura
// enchantment with enchant creature, read by its statics), `by` may pay
// `endCost` to end that effect, and it is attached to `host`. Reports
// whether the licid was on the battlefield to change; the caller checks it
// is still the object whose ability this is (CR 400.7).
//
// Caller must hold g.mu in write mode.
func (g *Game) BecomeLicidAuraForEffect(licid, by uuid.UUID, endCost string, host uuid.UUID) (bool, error) {
	if !g.BecomeAuraForEffect(licid) {
		return false, nil
	}
	c := findBattlefieldCard(g, licid)
	c.AuraFormEndBy = by
	c.AuraFormEndCost = endCost
	// A creature that was attached to something (CR 704.5p would have
	// unattached it) starts from nothing: the Aura is attached to the
	// target, and only to it.
	if c.IsAttached() {
		if err := g.UnattachForEffect(licid); err != nil {
			return true, err
		}
	}
	g.RecomputeLayersIfStaleLocked()
	return true, g.AttachForEffect(licid, TargetRef{Kind: TargetCard, ID: host})
}

// endEffectLocked is the end_effect performer: the effect ends, so the
// licid is a creature with its ability again, and it becomes unattached.
// Caller must hold g.mu.
func (g *Game) endEffectLocked(_ *Player, cardID uuid.UUID, _ SpecialAction) error {
	c := findBattlefieldCard(g, cardID)
	if c == nil {
		return ErrCardNotFound
	}
	c.ReturnedAura = false
	c.AuraFormEndBy = uuid.Nil
	c.AuraFormEndCost = ""
	c.effective = nil
	g.layerVersion.Add(1)
	if err := g.UnattachForEffect(cardID); err != nil {
		return err
	}
	g.RecomputeLayersIfStaleLocked()
	return nil
}
