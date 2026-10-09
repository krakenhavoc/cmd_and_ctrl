package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reiterating Bolt — Sorcery {1}{R}:
//
//	"Replicate—Pay {E}{E}{E}. (When you cast this spell, copy it for
//	 each time you paid its replicate cost. You may choose new targets
//	 for the copies.)
//	 Reiterating Bolt deals 3 damage to target creature or planeswalker."
//
// Replicate (CR 702.56a, ADR 0129 §5) is an optional cost keyed
// replicate, paid any number of times, with three energy per payment
// (ReplicatePayEnergy), and its cast trigger copying the spell once per
// payment with new targets offered (Replicate). The cast is refused when
// the payments times three exceed the caster's energy (CR 118.3), and
// the energy is never waived (ADR 0129 §4).
//
// The announcement is capped at ten payments, multikicker's reason (an
// announcement has to be finite): thirty energy is far past what any
// table reaches, and the stepper also stops at what the caster's energy
// pays for.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "371da0d8-3845-458e-9b0a-6477017b890e",
		Name:          "Reiterating Bolt",
		Completeness:  CompletenessFull,
		Targets:       TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		Purpose:       ForTargets(DamageToTarget(0, 3)),
		OptionalCosts: []game.AdditionalCost{ReplicatePayEnergy(3, 10)},
		Triggered:     []game.TriggeredAbility{Replicate()},
		OnResolve:     damageToFirstTarget(3),
	})
}
