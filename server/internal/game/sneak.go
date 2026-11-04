package game

import "github.com/google/uuid"

// sneak.go — #4586: sneak (CR 702.190, Marvel Super Heroes and TMNT).
//
//	"Sneak [cost]" means "Any time you could cast an instant during your
//	 declare blockers step, you may cast this spell by paying [cost] and
//	 returning an unblocked creature you control to its owner's hand
//	 rather than paying this spell's mana cost." (CR 702.190a)
//
//	"A permanent spell whose sneak cost was paid enters the battlefield
//	 tapped and attacking (see rule 506.3a). It will be attacking the same
//	 player, planeswalker, or battle as the creature that was returned
//	 to its owner's hand to pay the sneak cost …" (CR 702.190b)
//
// Ninjutsu's cast-time sibling, and built from ninjutsu's pieces rather
// than beside them. The price is an ordinary alternative cost (ADR 0025):
// ManaCost plus the ReturnToHand component Daze uses, narrowed by
// effects.UnblockedAttacker, the predicate ninjutsu's cost reads, so the
// view's picker, the enumerator and the validator share one candidate
// walk. AlternativeCost.Sneak adds the two things a price cannot say:
//
//   - the WINDOW. The offer opens instant-speed timing for its claim
//     (as AsThoughFlash does, CR 601.3c) and closes everywhere but its
//     caster's own declare blockers step. A per-player "only as a
//     sorcery" still shuts it, since the caster then could not cast an
//     instant. The cost already refuses an attacker nobody has finished
//     blocking (UnblockedAttackerForEffect, #1279), so "before blocks"
//     needs no rule of its own.
//   - the ENTRY. What the returned creature was attacking is gone the
//     moment the return clears Card.AttackingTarget, and LKI carries no
//     combat state, so the cast reads it first and keeps it on the
//     spell's PaidCost.ReturnedAttacking: the field ninjutsu's ability
//     already keeps it in, already in every snapshot. The resolving
//     permanent's entry event is seeded tapped and attacking that
//     defender, through stampEntryAttackerLocked, the door ninjutsu and
//     the attacking tokens share: put onto the battlefield attacking,
//     never declared (CR 508.4), and unblocked because it arrives in
//     the declare blockers step (CR 508.4d).
//
// A COPY of a sneaked permanent spell (#4646) copies the sneak cost with
// the rest of its alternative costs (CR 707.10: "all decisions made for
// it, including … additional or alternative costs"), so the token it
// becomes (CR 707.10f) enters tapped and attacking the same defender.
// createSpellCopyLocked carries AltCost and PaidCost.ReturnedAttacking;
// resolvePermanentSpellCopyLocked reads them through applySneakCopyEntry.
//
// A GRANTED sneak (#4680, Ninja Teen's "Creature cards in your graveyard
// have sneak {3}{B}. You may cast creature spells from your graveyard
// using their sneak abilities.") is Goblin Formula's shape: a standing
// CastPermission over the graveyard whose AltCostKey is
// AltCostKeyGrantedSneak and whose SneakCost is the price. The permission
// is the door and the synthesised offer is the price, so the card can be
// cast from there only with a sneak offer (validateCastPathLocked), its
// own or the granted one, and never for its mana cost.

// AltCostKeyGrantedSneak is the key of a granted sneak offer (#4680): the
// AltCostKey of the permission that grants it and the claim a cast makes.
// Its own key, not the printed "sneak", so a card that prints sneak keeps
// its own offer beside the granted one and the caster picks (CR 113.2c).
// An on-disk identity on StackItem.AltCost: never renamed, never reused.
const AltCostKeyGrantedSneak = "granted-sneak"

// SneakPayLabel is the card clause of a sneak cost as CR 702.190a words
// it: what the picker asks the caster to return.
const SneakPayLabel = "an unblocked creature you control"

// SneakReturnSpec is the "returning an unblocked creature you control to
// its owner's hand" half of a sneak cost (CR 702.190a): one creature the
// caster controls that is an unblocked attacker (UnblockedAttackerForEffect,
// ninjutsu's predicate). Shared by effects.Sneak and the granted offer, so
// a printed and a granted sneak take the same payment by construction.
// A cost, not a target (CR 601.2h): matched, so hexproof does not hide it.
func SneakReturnSpec() *TargetSpec {
	return &TargetSpec{
		Label: SneakPayLabel,
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(g *Game, caster uuid.UUID, c Card, _ ZoneKind) bool {
			return c.Controller == caster && c.IsCreature() && g.UnblockedAttackerForEffect(c.InstanceID)
		},
		Min: 1, Max: 1,
	}
}

// grantedSneakOffer is the offer a granted-sneak permission synthesises
// for `card` (CastPermission.AlternativeCostFor): its SneakCost, the
// return, and the Sneak bit that brings the window and the entry. Nil
// when the permission names no price, or the spell is not a creature
// spell ("creature spells", judged on the face being cast).
func grantedSneakOffer(p *CastPermission, card Card) *AlternativeCost {
	if p.SneakCost == "" || !card.IsCreature() {
		return nil
	}
	label := "Sneak " + p.SneakCost
	if p.Label != "" {
		label += " (" + p.Label + ")"
	}
	return &AlternativeCost{
		Key:            AltCostKeyGrantedSneak,
		Label:          label,
		ManaCost:       p.SneakCost,
		ReturnToHand:   SneakReturnSpec(),
		PayLabel:       SneakPayLabel,
		Sneak:          true,
		FromZone:       p.Zone,
		FromPermission: true,
	}
}

// sneakOnlyClaimOK is validateCastPathLocked's rule for a cast that a
// granted-sneak permission is the reason for (#4680): "You may cast
// creature spells from your graveyard using their sneak abilities". The
// spell must be a creature spell and the claim a sneak offer, the card's
// own or the granted one. Its mana cost, and any other alternative cost
// it prints, are not ways in.
func sneakOnlyClaimOK(card Card, alt *AlternativeCost) bool {
	return card.IsCreature() && alt != nil && alt.Sneak
}

// printedSneakClaim reports whether `claim` names a sneak offer `card`
// prints, so the cast that claims it looks for a granted-sneak permission
// to be cast under (castPermissionForClaimMatchingLocked).
func printedSneakClaim(card Card, claim string) bool {
	alt := AlternativeCostByKey(castOfferKey(card), claim)
	return alt != nil && alt.Sneak
}

// sneakWindowOpenLocked reports whether it is `playerID`'s own declare
// blockers step: the half of CR 702.190a's window that is not "any time
// you could cast an instant". Caller must hold g.mu.
func (g *Game) sneakWindowOpenLocked(playerID uuid.UUID) bool {
	return g.State == StateActive && g.Turn.Step == StepDeclareBlockers && g.activeSeatIDLocked() == playerID
}

// opensOwnCastWindow reports whether claiming this offer opens
// instant-speed timing for the cast: "as though it had flash"
// (AsThoughFlash, CR 601.3c) or sneak's "any time you could cast an
// instant" (CR 702.190a). Nil-safe.
func (a *AlternativeCost) opensOwnCastWindow() bool {
	return a != nil && (a.AsThoughFlash || a.Sneak)
}

// OpensOwnCastWindow is opensOwnCastWindow for the view's castable_here
// walk, which asks CastTimingForOfferOpenLocked only about the offers
// that could open a window the printed cast leaves shut.
func (a *AlternativeCost) OpensOwnCastWindow() bool { return a.opensOwnCastWindow() }

// sneakDefenderLocked is what the creature `returned` is attacking
// right now: the player, planeswalker or battle a sneak permanent will
// attack (CR 702.190b). Read before the return, because the exit clears
// it. Caller must hold g.mu.
func (g *Game) sneakDefenderLocked(returned uuid.UUID) uuid.UUID {
	if c := findBattlefieldCard(g, returned); c != nil {
		return c.AttackingTarget
	}
	return uuid.Nil
}

// applySneakEntryLocked seeds CR 702.190b onto a resolving permanent
// spell's entry event: tapped, and attacking what the returned creature
// was attacking. Seeded before the CR 614 pipeline, like Tapped and
// Attacking on every other entry, so a pause-and-resume carries it.
//
// CR 506.3a: a permanent that is not a creature still enters (tapped),
// but is never an attacking permanent, so it is not stamped attacking.
// A defender that has since become unattackable leaves the creature on
// the battlefield not attacking (CR 506.3c), stampEntryAttackerLocked's
// posture.
//
// No-op for a spell cast any other way. Caller must hold g.mu.
func (g *Game) applySneakEntryLocked(ev *ReplacementEvent, card Card, item *StackItem) {
	if ev == nil || !sneakPaid(card, item) {
		return
	}
	ev.EntersTapped = true
	if card.IsCreature() {
		ev.EntersAttacking = item.Paid.ReturnedAttacking
	}
}

// sneakPaid reports whether `item` is a spell whose sneak cost was paid
// (CR 702.190b): its AltCost names a sneak offer of `card`'s catalog row,
// or is a granted sneak (#4680), whose key no catalog row prints. True for
// a copy of such a spell too, since a copy copies "additional or
// alternative costs" paid (CR 707.10). Nil-safe.
func sneakPaid(card Card, item *StackItem) bool {
	if item == nil || item.AltCost == "" {
		return false
	}
	if item.AltCost == AltCostKeyGrantedSneak {
		return true
	}
	alt := AlternativeCostByKey(CatalogKey(card), item.AltCost)
	return alt != nil && alt.Sneak
}

// applySneakCopyEntry is applySneakEntryLocked for the token a resolving
// copy of a permanent spell becomes (CR 707.10f, 608.3f). The copy copied
// the sneak cost the original paid (CR 707.10), so the token enters tapped
// and attacking what the returned creature was attacking: CR 702.190b
// says "a permanent spell whose sneak cost was paid", and a copy of a
// permanent spell is one. `tmpl` is the token's template, which decides
// the creature half (CR 506.3a). It sets entry.Tapped and returns the
// defender for the token's group; uuid.Nil leaves it not attacking.
// No-op for any other copy.
func applySneakCopyEntry(entry *TokenEntryOptions, tmpl Card, item *StackItem) (attacking uuid.UUID) {
	if entry == nil || !sneakPaid(tmpl, item) {
		return uuid.Nil
	}
	entry.Tapped = true
	if tmpl.IsCreature() {
		return item.Paid.ReturnedAttacking
	}
	return uuid.Nil
}
