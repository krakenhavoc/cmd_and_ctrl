package game

import (
	"slices"

	"github.com/google/uuid"
)

// attach_graveyard.go is ADR 0154 (#1780): an Aura attached to a card
// that is not on the battlefield, and the enchant clause Animate Dead
// gains as it enters.
//
// CR 303.4 lets an Aura enchant any object its enchant ability allows,
// and "enchant creature card in a graveyard" allows a card. The
// relation is still Card.AttachedTo; what this file adds is
//
//   - the door: AttachForEffect accepts an off-battlefield host only
//     when the Aura's own enchant clause lists the host's zone and
//     accepts the host (offBattlefieldHostLocked), and records the
//     host's CR 400.7 epoch in Card.AttachedEpoch;
//   - the swapped clause: Card.EnchantSwapped and Card.EnchantReturned,
//     "enchant creature put onto the battlefield with this Aura", read
//     first by attachmentLegalLocked and by AttachForEffect
//     (swappedEnchantAdmitsLocked).
//
// Everything else an attachment does is unchanged: AttachedHostOf,
// AttachmentsOf and every "enchanted creature" static read the
// battlefield only, so a creature CARD in a graveyard is never
// "enchanted creature" (ADR 0154 decision 10).

// offBattlefieldHostLocked reports whether `attachmentID` may be attached
// to the card `host` names although that card is not on the battlefield
// (ADR 0154 decision 2), and returns the host's object epoch when it
// may. All three must hold:
//
//   - the attachment is an Aura (an Equipment or Fortification attaches
//     only to a permanent, CR 301.5c and 301.6),
//   - its catalogued enchant clause lists the zone the host is in, and
//   - that clause accepts the host, without the targeting gate
//     (attaching does not target).
//
// Anything else is the caller's quiet CR 701.3b skip. A swapped Aura is
// never admitted here: its clause names a permanent.
//
// Caller must hold g.mu.
func (g *Game) offBattlefieldHostLocked(attachmentID uuid.UUID, host TargetRef) (int, bool) {
	c := findBattlefieldCard(g, attachmentID)
	if c == nil || !c.IsAura() || c.EnchantSwapped || c.Bestowed {
		return 0, false
	}
	spec := TargetSpecFor(catalogKeyOf(c))
	if spec == nil {
		return 0, false
	}
	z := g.findCardZoneLocked(host.ID)
	if z == nil || z.Kind == ZoneBattlefield || z.Kind == ZoneStack {
		return 0, false
	}
	if !slices.Contains(spec.Zones, z.Kind) {
		return 0, false
	}
	if !g.specMatchLocked(SourceChooser(c.Controller), spec, host, false) {
		return 0, false
	}
	return g.cardObjectEpochLocked(host.ID), true
}

// swappedEnchantAdmitsLocked is "enchant creature put onto the
// battlefield with this Aura" (ADR 0154 decision 4) for one host: the
// permanent EnchantReturned names, while it is that object (CR 400.7)
// and a creature (effective types). Protection is not asked here:
// attachmentLegalLocked asks it of every attachment before it reaches
// the clause, and an attach onto a protected creature happens and is
// undone by the state-based action, which is Animate Dead's ruling.
//
// A zero EnchantReturned (nothing came back) admits nothing.
//
// Caller must hold g.mu. Must run after the layer recompute when it is
// judging legality, like every type read in attachmentLegalLocked.
func (g *Game) swappedEnchantAdmitsLocked(c *Card, host TargetRef) bool {
	want := c.EnchantReturned
	if want.ID == uuid.Nil || host.Kind != TargetCard || host.ID != want.ID {
		return false
	}
	h := findBattlefieldCard(g, host.ID)
	return h != nil && h.ObjectEpoch == want.Epoch && h.IsCreature()
}

// EnchantedGraveyardCardForEffect is "enchanted creature card" for an
// Aura that enchants a card in a graveyard (ADR 0154 decision 5, step
// 3): the card the Aura on the battlefield is attached to, while that
// card is in a graveyard and is still the object the Aura was attached
// to (CR 400.7, Card.AttachedEpoch). False otherwise, which is "there is
// no enchanted creature card".
//
// Caller must hold g.mu.
func (g *Game) EnchantedGraveyardCardForEffect(auraID uuid.UUID) (uuid.UUID, bool) {
	aura := findBattlefieldCard(g, auraID)
	if aura == nil || aura.AttachedTo.Kind != TargetCard {
		return uuid.Nil, false
	}
	id := aura.AttachedTo.ID
	z := g.findCardZoneLocked(id)
	if z == nil || z.Kind != ZoneGraveyard || g.cardObjectEpochLocked(id) != aura.AttachedEpoch {
		return uuid.Nil, false
	}
	return id, true
}

// SwapEnchantClauseForEffect is "it loses 'enchant creature card in a
// graveyard' and gains 'enchant creature put onto the battlefield with
// this Aura'" (ADR 0154 decision 4) on the Aura on the battlefield.
// The creature is named afterwards, by SetEnchantReturnedForEffect, once
// it is on the battlefield; until then the clause admits nothing.
// Reports whether the Aura was there to change.
//
// Caller must hold g.mu.
func (g *Game) SwapEnchantClauseForEffect(auraID uuid.UUID) bool {
	aura := findBattlefieldCard(g, auraID)
	if aura == nil {
		return false
	}
	aura.EnchantSwapped = true
	aura.EnchantReturned = ObjectRef{}
	return true
}

// SetEnchantReturnedForEffect names the creature a swapped Aura's
// clause admits: the permanent the same resolution put onto the
// battlefield (ADR 0154 decision 5, step 5). Reports whether the Aura
// is on the battlefield with the swapped clause.
//
// Caller must hold g.mu.
func (g *Game) SetEnchantReturnedForEffect(auraID uuid.UUID, ref ObjectRef) bool {
	aura := findBattlefieldCard(g, auraID)
	if aura == nil || !aura.EnchantSwapped {
		return false
	}
	aura.EnchantReturned = ref
	return true
}
