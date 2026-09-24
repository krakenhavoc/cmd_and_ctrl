package game

import "github.com/google/uuid"

// return_cost.go — #1213: returning permanents you control to their
// owners' hands as part of an ability's cost (CR 602.2, CR 118.3).
//
// Quirion Ranger's "Return a Forest you control to its owner's hand",
// Wirewood Symbiote's Elf, Master Transmuter's "{U}, {T}, Return an
// artifact you control to its owner's hand", Meloku the Clouded
// Mirror's "{1}, Return a land you control to its owner's hand".
//
// It is TapOthersCost (#758) one verb over, and deliberately built on
// the same five pieces — the struct, one options walk, one payability
// predicate, one validator, one payer — so the engine, the protocol
// view's picker and the legal-move enumerator cannot disagree about
// which permanent pays (#544).
//
// Two things it is NOT, and both distinctions are observable:
//
//   - **Not AbilityCost.SacrificeOther with a destination.** A
//     sacrifice goes through sacrificePermanentLocked: EventSacrifice,
//     a dies-trigger, CR 701.17a. A return is an ordinary battlefield
//     exit to a hand, and Blood Artist must not fire on one.
//
//   - **Not BounceToHandForEffect.** That is the EFFECT verb (Azorius
//     Chancery's trigger, Chain of Vapor). Paying a cost is not an
//     effect, and CR 601.2h / 602.2b pay an announcement's costs as
//     ONE INDIVISIBLE STEP — so the move may not stop on a CR 903.9
//     prompt with the ability half announced. The payer sets
//     zoneRoute.MustSettleNow, which is the same bit, set for the same
//     reason, that a discard paid as a cost sets. The CR 903.9 question
//     is asked before the payment instead (#1397,
//     cost_commander_choice.go).
//
// And one it shares with every other cost component: paying a cost
// does not TARGET (CR 601.2h / 602.2b), so the filter is matched with
// specMatchLocked(…, false) and shroud, hexproof and protection never
// apply — exactly as tap_others_cost.go, counter_cost.go and
// activated.go's SacrificeOther match theirs.

// ReturnToHandCost is the "return N [permanents] you control to their
// owners' hands" component of an ability's cost.
//
// The zero value, and nil, demand nothing — so the activation path can
// ask without a guard, the way it asks TapOthersCost.
type ReturnToHandCost struct {
	// Count is how many permanents must be returned. 1 for every card
	// that prints the clause today.
	//
	// Fixed, for the reason TapOthersCost.Count is: a VARIABLE count
	// would need the announce path SacrificeCostBounds uses, and no
	// card prints one.
	Count int

	// Filter is what may be returned ("a Forest you control", "an
	// artifact you control"), reused from the TargetSpec vocabulary
	// exactly as AbilityCost.SacrificeOther and TapOthersCost.Filter
	// reuse it.
	//
	// It does NOT target. The "you control" clause is enforced by the
	// validator below rather than asked of the spec, so a filter that
	// forgot it cannot return an opponent's permanent.
	//
	// Nil is not a cost: Empty reports true, because "return nothing
	// in particular" is not a printed clause.
	Filter *TargetSpec

	// ExcludeSource is the printed word "another". No card prints it
	// on this clause yet; the field exists because "another" is a
	// property of the printed cost rather than of the engine, and
	// hard-coding either reading is what TapOthersCost.ExcludeSource
	// exists to avoid.
	//
	// When it is CLEAR the source may pay if the filter admits it —
	// Master Transmuter is an artifact you control and returning
	// herself is a normal, sometimes correct, line.
	ExcludeSource bool

	// Label is the clause as printed, shown above the client's picker
	// so the prompt reads like the card. "a Forest you control", not
	// "choose 1".
	Label string
}

// Empty reports whether the cost demands nothing. Nil-safe, so the
// activation path can ask without a guard.
func (c *ReturnToHandCost) Empty() bool {
	return c == nil || c.Count < 1 || c.Filter == nil
}

// ReturnToHandOptionsForEffect is the set of permanents that could pay
// `rc` for an activation of an ability on `sourceID` by `playerID`, in
// battlefield order. Empty when nothing on the board matches.
//
// ONE walk, shared by the protocol view (the client's picker), the
// legal-move enumerator (the bots) and — through the same predicates —
// validateReturnToHandCostLocked, so the three cannot disagree about
// which permanent pays. That is the #544 invariant and the reason
// tap_others_cost.go has exactly one of these too.
//
// It is the NON-targeting candidate walk (specCandidatesLocked):
// choosing a permanent to pay a cost does not target it, so a hexproof
// creature is still a legal return.
//
// Whether the options add up to Count is the PAYMENT's question,
// answered by ReturnToHandPayable and by the validator. This walk
// never enumerates subsets — that is the combinatorial expansion #544
// warns about.
//
// Caller must hold g.mu (read or write).
func (g *Game) ReturnToHandOptionsForEffect(playerID, sourceID uuid.UUID, rc *ReturnToHandCost) []uuid.UUID {
	if rc.Empty() {
		return nil
	}
	var out []uuid.UUID
	for _, id := range g.specCandidatesLocked(playerID, rc.Filter).Cards {
		if rc.ExcludeSource && id == sourceID {
			continue
		}
		c := findBattlefieldCard(g, id)
		if c == nil || c.Controller != playerID {
			continue
		}
		out = append(out, id)
	}
	return out
}

// ReturnToHandPayable reports whether `rc` could be paid at all right
// now — the question the legal enumerator and the client's greyed menu
// row ask before offering the activation, and the one CR 118.3 turns
// into a hard refusal.
//
// Caller must hold g.mu (read or write).
func (g *Game) ReturnToHandPayable(playerID, sourceID uuid.UUID, rc *ReturnToHandCost) bool {
	if rc.Empty() {
		return true
	}
	return len(g.ReturnToHandOptionsForEffect(playerID, sourceID, rc)) >= rc.Count
}

// validateReturnToHandCostLocked checks that every permanent the
// activator named may actually pay this cost, without moving anything
// — ADR 0020 §3's "validate everything, then pay everything", so a
// refused activation never leaves a board half returned.
//
// What it enforces, in the order the errors matter:
//
//   - IDs sent for an ability with no return component are rejected
//     rather than ignored, as tap_ids, crew_ids and sacrifice_ids are:
//     a client that sends them is confused about which ability it is
//     firing.
//   - Exactly Count permanents, no more and no fewer. A cost is not
//     partially payable (CR 118.3), and over-naming would bounce a
//     permanent for nothing.
//   - No permanent named twice — otherwise one card would pay both
//     halves of a two-permanent clause.
//   - The source itself when ExcludeSource is set ("another").
//   - On the battlefield (ErrCardNotFound), controlled by the
//     activator (ErrCardCallerMismatch) and matched by Filter WITHOUT
//     the targeting gate (ErrIllegalTarget).
//
// A permanent that is TAPPED is a perfectly legal pick: returning is
// not tapping, and nothing about CR 118.3 says otherwise. That is the
// one line tap_others_cost.go's validator has and this one does not.
//
// Caller must hold g.mu.
func (g *Game) validateReturnToHandCostLocked(playerID, sourceID uuid.UUID, rc *ReturnToHandCost, ids []uuid.UUID) error {
	if rc.Empty() {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	if len(ids) != rc.Count {
		return ErrInvalidParam
	}
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			return ErrInvalidParam
		}
		seen[id] = true
		if rc.ExcludeSource && id == sourceID {
			return ErrInvalidParam
		}
		c := findBattlefieldCard(g, id)
		if c == nil {
			return ErrCardNotFound
		}
		// "You control" is a cost's own clause, enforced here rather
		// than asked of the filter, so a filter that forgot it cannot
		// spend an opponent's board.
		if c.Controller != playerID {
			return ErrCardCallerMismatch
		}
		// specMatchLocked(…, false), not targetLegalLocked: returning
		// a permanent to pay a cost does not target it (CR 601.2h /
		// 602.2b), so the CR 702 keyword gate must not apply.
		if !g.specMatchLocked(SourceChooser(playerID), rc.Filter, TargetRef{Kind: TargetCard, ID: id}, false) {
			return ErrIllegalTarget
		}
	}
	return nil
}

// payReturnToHandCostLocked moves the validated permanents to their
// OWNERS' hands (CR 109.5 — "its owner's hand", never the activator's),
// one route each, so a leaves-the-battlefield watcher sees every one of
// them.
//
// Returns what the first returned permanent that was ATTACKING was
// attacking, for PaidCost.ReturnedAttacking (#1227). It is read HERE,
// before the bounce, because it cannot be read anywhere else: the exit
// clears Card.AttackingTarget (zone.go) and LKI carries no combat
// state, so ninjutsu's "attacking the same player or planeswalker that
// the returned creature was attacking" (CR 702.49a) would be
// unanswerable one line later. uuid.Nil for every other printed return
// cost, whose permanent is a Forest or an artifact and is attacking
// nothing.
//
// Through the ONE exit door (routeCardToZoneLocked) rather than through
// BounceToHandForEffect, and with MustSettleNow set: CR 601.2h /
// 602.2b pay an announcement's costs as one indivisible step, so the
// move may not stop on a prompt. A commander returned this way is still
// offered CR 903.9 — its owner was asked BEFORE the payment began
// (askCostCommanderLocked, #1397), and `answers` carries what they
// said onto the move.
//
// Call only after validateReturnToHandCostLocked has passed, and —
// as with payCostSacrificesLocked — BEFORE the ability's stack item is
// built, so the leaves-triggers this queues are drained by the closing
// runStateChecksLocked and sit ABOVE the ability (CR 603.3b).
//
// A permanent that vanished between validation and payment is skipped
// rather than erroring: nothing can happen between the two under one
// lock today, and a partial board is still better than a panic if that
// ever stops being true.
//
// Caller must hold g.mu.
func (g *Game) payReturnToHandCostLocked(playerID, sourceID uuid.UUID, ids []uuid.UUID, answers map[uuid.UUID]bool) (uuid.UUID, error) {
	var attacking uuid.UUID
	for _, id := range ids {
		c := findBattlefieldCard(g, id)
		if c == nil {
			continue
		}
		if attacking == uuid.Nil {
			attacking = c.AttackingTarget
		}
		if _, err := g.routeCardToZoneLocked(zoneRoute{
			CardID:        id,
			Dst:           ZoneHand,
			Actor:         playerID,
			Source:        sourceID,
			Cause:         MoveCause{Kind: MoveCauseCost, Controller: playerID},
			MustSettleNow: true,
			// #1397: the owner's CR 903.9 answer, asked before the
			// payment by askCostCommanderLocked.
			commanderAnswer: commanderAnswerFor(answers, id),
		}); err != nil {
			return attacking, err
		}
	}
	return attacking, nil
}
