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
	if ev == nil || item == nil || item.AltCost == "" {
		return
	}
	alt := AlternativeCostByKey(CatalogKey(card), item.AltCost)
	if alt == nil || !alt.Sneak {
		return
	}
	ev.EntersTapped = true
	if card.IsCreature() {
		ev.EntersAttacking = item.Paid.ReturnedAttacking
	}
}
