package game

import "github.com/google/uuid"

// beam_me_up.go: beam me up (Star Trek, #4973).
//
//	"Beam me up [cost] (You may cast this card from your graveyard for
//	 [cost] if you also return a creature you control to its owner's
//	 hand. Then exile this spell.)"
//
// The pinned Comprehensive Rules (2026-09-25) have no entry for it yet,
// so the reminder text governs. It is flashback's shape (CR 702.34a)
// with one more payment: an alternative cost from the graveyard, priced
// at [cost] plus the ReturnToHand component Daze and sneak already use,
// and exiled as it leaves the stack. The printed keyword is
// effects.BeamMeUp; this file holds what the printed keyword, Scotty,
// Miracle Worker's granted one and Marooned share:
//
//   - BeamMeUpReturnSpec, the return. One creature the caster controls,
//     matched and never targeted (CR 601.2h), so hexproof does not hide
//     it, and never one that "can't be beamed up".
//   - Characteristic.CantBeBeamedUp, Marooned's "Enchanted permanent ...
//     can't be beamed up". A permanent is "beamed up" when it is
//     returned to pay a beam me up cost (Scotty's own text: "Scotty
//     can't be beamed up to cast a spell this way"), so the flag only
//     narrows that return. Written by an ordinary layer-6 static of the
//     restricting Aura, Blossombind's CantHaveCounters shape, so it ends
//     when the Aura leaves and a "loses all abilities" on the host does
//     not lift it.
//   - AltCostKeyGrantedBeamMeUp, Scotty's "each card in your graveyard
//     has beam me up. The beam me up cost is equal to that spell's mana
//     cost." A standing CastPermission under this key, whose offer
//     AlternativeCostFor builds: the card's mana cost, the return with
//     the granting permanent left out, and the exile.

// AltCostKeyBeamMeUp is the printed keyword's offer key, on the wire as
// `alternative_cost` and on StackItem.AltCost. An on-disk identity:
// never renamed, never reused.
const AltCostKeyBeamMeUp = "beam-me-up"

// AltCostKeyGrantedBeamMeUp is the key of a granted beam me up offer
// (Scotty, Miracle Worker): the AltCostKey of the permission and the
// claim a cast makes. Its own key, so a card that prints beam me up keeps
// its own offer beside the granted one and the caster picks (CR 113.2c).
// An on-disk identity: never renamed, never reused.
const AltCostKeyGrantedBeamMeUp = "granted-beam-me-up"

// BeamMeUpPayLabel is the card clause of a beam me up cost as the
// reminder words it: what the picker asks the caster to return.
const BeamMeUpPayLabel = "a creature you control"

// BeamMeUpReturnSpec is the "if you also return a creature you control to
// its owner's hand" half of a beam me up cost: one creature the caster
// controls that can be beamed up, and never `except` (Scotty's "Scotty
// can't be beamed up to cast a spell this way"; uuid.Nil excludes
// nothing). A cost, not a target (CR 601.2h).
func BeamMeUpReturnSpec(except uuid.UUID) *TargetSpec {
	label := BeamMeUpPayLabel
	if except != uuid.Nil {
		label = "another creature you control"
	}
	return &TargetSpec{
		Label: label,
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(g *Game, caster uuid.UUID, c Card, _ ZoneKind) bool {
			return c.Controller == caster && c.IsCreature() && c.InstanceID != except &&
				!g.cantBeBeamedUpLocked(findBattlefieldCard(g, c.InstanceID))
		},
		Min: 1, Max: 1,
	}
}

// cantBeBeamedUpLocked reports whether the permanent `c` can't be
// returned to pay a beam me up cost: its effective characteristics say
// so. False for nil. Caller must hold g.mu.
func (g *Game) cantBeBeamedUpLocked(c *Card) bool {
	return c != nil && c.Effective().CantBeBeamedUp
}

// CantBeBeamedUpForEffect is cantBeBeamedUpLocked by instance ID, for
// card files and tests. False for a card not on the battlefield. Caller
// must hold g.mu.
func (g *Game) CantBeBeamedUpForEffect(cardID uuid.UUID) bool {
	return g.cantBeBeamedUpLocked(findBattlefieldCard(g, cardID))
}

// grantedBeamMeUpOffer is the offer a granted beam me up permission
// synthesises for `card` (CastPermission.AlternativeCostFor): the card's
// own mana cost, the return of a creature other than the permission's
// source, and the exile as it leaves the stack. Nil for a card that is
// not an instant or a sorcery (the reminder: "Only instant and sorcery
// spells can be cast this way"), judged on the face being cast.
func grantedBeamMeUpOffer(p *CastPermission, card Card) *AlternativeCost {
	if !card.IsInstant() && !card.IsSorcery() {
		return nil
	}
	label := "Beam me up — its mana cost"
	if p.Label != "" {
		label += " (" + p.Label + ")"
	}
	return &AlternativeCost{
		Key:                 AltCostKeyGrantedBeamMeUp,
		Label:               label,
		ManaCost:            card.ManaCost,
		ReturnToHand:        BeamMeUpReturnSpec(p.Source),
		PayLabel:            "another creature you control",
		FromZone:            p.Zone,
		FromPermission:      true,
		ExileOnLeavingStack: true,
	}
}
