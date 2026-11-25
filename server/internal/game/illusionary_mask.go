package game

import (
	"github.com/google/uuid"
)

// illusionary_mask.go — #6219: Illusionary Mask.
//
//	{X}: You may choose a creature card in your hand whose mana cost
//	could be paid by some amount of, or all of, the mana you spent on
//	{X}. If you do, you may cast that card face down as a 2/2 creature
//	spell without paying its mana cost. If the creature that spell
//	becomes as it resolves has not been turned face up and would assign
//	or deal damage, be dealt damage, or become tapped, instead it's
//	turned face up and assigns or deals damage, is dealt damage, or
//	becomes tapped. Activate only as a sorcery.
//
// CR 708.2: "Face-down spells and face-down permanents have no
// characteristics other than those listed by the ability or rules that
// allowed the spell or permanent to be face down." The Mask lists a 2/2
// creature, the body every face-down permanent state has, so the spell is
// an ordinary CR 708.2 object of its own kind, FaceDownMasked:
//
//   - The cast is a resolution cast (CR 608.2g, resolution_copy_cast.go)
//     of a card in hand under a permission whose offer is the free
//     face-down cast (AltCostKeyGrantedFaceDownFree, AlternativeCostFor),
//     so the price, the timing and the face-down stamping are the morph
//     cast's (ADR 0082). Choosing the card and casting it are one answer:
//     choosing without casting changes nothing.
//   - The candidates are the creature cards in hand whose mana cost the
//     mana the ability spent could pay (MaskCandidatesForEffect).
//   - The kind is the replacement's marker. It is a permanent state like
//     `turned` (an effect, not a keyword, put it face down), so only the
//     card's own morph or disguise turns it up voluntarily (CR 708.7,
//     TurnFaceUpOffers), and turning it face up by any route ends it: the
//     text's "has not been turned face up".
//   - "Instead it's turned face up and …" is applied at the three places
//     the engine does those things: every tap emits EventTapCard and an
//     attack's tap is locked in by commitAttackDeclarationLocked; every
//     damage event passes damageThroughReplacementsLocked; and a combat
//     damage step snapshots what each combatant assigns in
//     assignAndDealCombatDamageLocked. Each turns a masked permanent face
//     up first (unmaskLocked), so the event happens to the face-up
//     creature. A permanent that can't be turned face up (CR 101.2,
//     Karlov Watchdog) stays face down and the event goes on.

// FaceDownMasked is Illusionary Mask's face-down spell and the permanent
// it becomes (#6219): a CR 708.2 2/2 with no text, turned face up instead
// of assigning or dealing damage, being dealt damage, or becoming tapped.
const FaceDownMasked FaceDownKind = "masked"

// AltCostKeyGrantedFaceDownFree is the claim of a permission that lets a
// card in hand be cast face down as a 2/2 creature spell without paying
// its mana cost (Illusionary Mask). An on-disk identity: never renamed.
const AltCostKeyGrantedFaceDownFree = "granted-face-down-free"

// grantedFaceDownFreeOffer is the offer that claim prices: no mana, and
// the CR 708.2 masked object. A land is never offered (AlternativeCostFor).
func grantedFaceDownFreeOffer(p *CastPermission) *AlternativeCost {
	label := p.Label
	if label == "" {
		label = "Cast it face down"
	}
	return &AlternativeCost{
		Key:            AltCostKeyGrantedFaceDownFree,
		Label:          label,
		FromZone:       p.Zone,
		FromPermission: true,
		FaceDown:       &FaceDownCast{Kind: FaceDownMasked},
	}
}

// MaskCandidatesForEffect lists the creature cards in `player`'s hand
// whose mana cost could be paid by some amount of, or all of, the mana
// `spent` (Illusionary Mask: the mana spent on {X}). A {X} in a card's
// own cost counts as 0. Caller must hold g.mu.
func (g *Game) MaskCandidatesForEffect(player uuid.UUID, spent []ManaToken) []uuid.UUID {
	p := g.playerByIDLocked(player)
	if p == nil || p.Hand == nil {
		return nil
	}
	pool := ManaPool(append([]ManaToken(nil), spent...))
	var out []uuid.UUID
	for _, c := range p.Hand.Cards {
		if !c.PrintedIsCreature() {
			continue
		}
		cost, err := ParseCost(c.ManaCost)
		if err != nil {
			continue
		}
		if pool.CanPay(cost, 0) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// OfferMaskCastForEffect is Illusionary Mask's resolution once its
// candidates are known: `chooser` may cast one of `cards` from their hand
// face down as a 2/2 creature spell without paying its mana cost, during
// this resolution (CR 608.2g). Returns false when nothing was offered (no
// candidate, or no resolution is open). Caller must hold g.mu (write).
func (g *Game) OfferMaskCastForEffect(chooser, source uuid.UUID, reason string, cards []uuid.UUID) bool {
	if len(cards) == 0 {
		return false
	}
	perm := g.FreeCastFromHandPermission(chooser, source, reason)
	perm.Cost = ""
	perm.AltCostKey = AltCostKeyGrantedFaceDownFree
	perm.Label = reason + " — cast it face down as a 2/2"
	return g.OfferResolutionCastForEffect(ResolutionCastOffer{
		Chooser:      chooser,
		Source:       source,
		Reason:       reason,
		Cards:        cards,
		Keyword:      MayCastKeywordHandCast,
		AcceptLabel:  "Cast it face down",
		DeclineLabel: "Keep it in your hand",
		Permission:   &perm,
	})
}

// unmaskOnTap is EmitEvent's hook for EventTapCard. Set in init, not at
// declaration: EmitEvent is reachable from the package's own variable
// initialisers, and a direct call would make an initialisation cycle.
var unmaskOnTap func(g *Game, cardID uuid.UUID)

func init() {
	unmaskOnTap = func(g *Game, cardID uuid.UUID) { g.unmaskLocked(cardID) }
}

// unmaskLocked is the Mask's "instead it's turned face up": `cardID`, if
// it is a masked face-down permanent, is turned face up first (CR 708.8;
// no cost, no new object). One that can't be turned face up (CR 101.2)
// stays face down. Caller must hold g.mu (write).
func (g *Game) unmaskLocked(cardID uuid.UUID) {
	if cardID == uuid.Nil || g.Battlefield == nil {
		return
	}
	c := findBattlefieldCard(g, cardID)
	if c == nil || !c.FaceDown || c.FaceDownKind != FaceDownMasked {
		return
	}
	if g.TurnFaceUpForEffect(uuid.Nil, c.Controller, cardID) {
		g.RecomputeLayersIfStaleLocked()
	}
}

// unmaskForDamageLocked turns a masked source or recipient of the damage
// event face up before it is dealt. Caller must hold g.mu (write).
func (g *Game) unmaskForDamageLocked(ev *ReplacementEvent) {
	if ev == nil || ev.Kind != RepEventDamage || ev.DamageAmount <= 0 {
		return
	}
	g.unmaskLocked(ev.DamageSource)
	g.unmaskLocked(ev.DamageTarget)
}

// unmaskCombatantsLocked turns face up every masked combatant that would
// assign combat damage in this step (CR 510.1), before what each assigns
// is read. Caller must hold g.mu (write).
func (g *Game) unmaskCombatantsLocked(firstStrike bool) {
	var ids []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !c.FaceDown || c.FaceDownKind != FaceDownMasked {
			continue
		}
		if c.AttackingTarget == uuid.Nil && c.BlockingTarget == uuid.Nil {
			continue
		}
		if !g.participatesInStepLocked(c, firstStrike) || g.permanentAssignsNoCombatDamageLocked(c) || c.combatDamagePower() <= 0 {
			continue
		}
		ids = append(ids, c.InstanceID)
	}
	for _, id := range ids {
		g.unmaskLocked(id)
	}
}
