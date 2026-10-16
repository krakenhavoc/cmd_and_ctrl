package game

import "github.com/google/uuid"

// tap_additional_cost.go — #3868: "As an additional cost to cast this
// spell, tap an untapped Vampire you control" (New Blood; CR 601.2h).
//
// The clause is the TapOthersCost the activated abilities (Earthcraft,
// station) and the alternative costs (Orim's Cure, Zahid) already carry,
// on the MANDATORY slot of a cast (AdditionalCost.TapOthers), so the
// candidate walk (TapOthersOptionsForEffect), the payability predicate
// (TapOthersPayable), the validator and the payer are theirs, and the
// view, the enumerator and the engine cannot disagree about which
// permanent pays (#544).
//
// What the cast path adds is only where the picks ride and when they are
// paid:
//
//   - On CastSpellParams.TeamworkIDs, the list Collective Effort's
//     escalate taps already use: both are "tap permanents you control as
//     a cost", neither is the {T} symbol, and effects.Register keeps a
//     card from carrying two of them, so the list is never ambiguous.
//     The auto-tapper already keeps those permanents out of the mana it
//     plans (CastAutoTapExclusions), so a Vampire named to New Blood is
//     not also tapped for its {B}.
//   - Validated with every other component before anything is paid
//     (ADR 0020 §3), and kept apart from the convoke / waterbend taps of
//     the same cast, so one creature cannot pay twice (CR 118.3).
//   - Tapped with the spell already on the stack (CR 601.2a before
//     601.2h), one EventTapCard each, so a "becomes tapped" watcher goes
//     above it; and recorded on PaidCost.TappedOthers, the record station
//     reads, so a resolution can name "the creature tapped to pay this
//     spell's additional cost" (Swallow Whole).
//
// #3961 adds the "any number" form (TapOthersCost.AnyNumber, Burn at the
// Stake, Explosive Singularity): the same list, validator and payer, with
// the count the caster's from zero up. None is a legal payment, so the
// cost never makes a spell uncastable, and the count is announced before
// the total is settled (CR 601.2b before 601.2f), so the pricer reads it
// as CostQuery.TappingToPay.
//
// No summoning-sickness check, deliberately: CR 302.6 hangs sickness on
// a creature's own {T} abilities, and this is a cost that taps a
// permanent, not the {T} symbol. A Vampire that entered this turn can
// pay for New Blood. Paying a cost does not target (CR 601.2h), so a
// hexproof Vampire pays too.

// planTapOthers is the tap-permanents component the announced plan
// demands, or nil. effects.Register allows it only on the mandatory
// slot, so there is at most one.
func planTapOthers(plan []costPayment) *TapOthersCost {
	for _, pay := range plan {
		if !pay.cost.TapOthers.Empty() {
			return pay.cost.TapOthers
		}
	}
	return nil
}

// validateCastTapOthersLocked checks the permanents named to pay the
// plan's tap-permanents cost: the activation's own validator (exactly the
// clause's count, distinct, on the battlefield, the caster's, untapped,
// matching the filter without the targeting gate), plus none of them also
// named to the cast's convoke / waterbend taps. With no such component in
// the plan it accepts anything: the list then belongs to teamwork or
// escalate, whose validator reads it.
//
// Caller must hold g.mu.
func (g *Game) validateCastTapOthersLocked(playerID, castID uuid.UUID, plan []costPayment, ids, otherTaps []uuid.UUID) error {
	tc := planTapOthers(plan)
	if tc == nil {
		return nil
	}
	tapping := make(map[uuid.UUID]bool, len(otherTaps))
	for _, id := range otherTaps {
		tapping[id] = true
	}
	for _, id := range ids {
		if tapping[id] {
			return ErrInvalidParam
		}
	}
	return g.validateTapOthersCostLocked(playerID, castID, tc, ids)
}

// payCastTapOthersLocked taps the validated permanents and records them
// on the spell's payment, AS THEY WERE TAPPED (paidTapsFrom), so the
// resolution can find the same object (CR 400.7) and a leave-the-
// battlefield freeze keeps its last-known power (CR 608.2h). Reports
// whether the plan had the component, so the caller pays teamwork's taps
// otherwise.
//
// Call only after validateCastTapOthersLocked has passed and with the
// spell on the stack. Caller must hold g.mu.
func (g *Game) payCastTapOthersLocked(playerID, castID uuid.UUID, plan []costPayment, ids []uuid.UUID) bool {
	if planTapOthers(plan) == nil {
		return false
	}
	g.RecomputeLayersIfStaleLocked()
	paid := g.payTapOthersCostLocked(playerID, ids)
	if item := g.StackMeta[castID]; item != nil {
		item.Paid.TappedOthers = paidTapsFrom(paid)
	}
	return true
}
