package effects

// Fling — Instant {1}{R}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Fling deals damage equal to the sacrificed creature's power to any
//	 target."
//
// The sacrifice is the cast's additional cost (CR 601.2b, 601.2h):
// exactly one creature, named as the spell is cast, and gone before
// anyone can respond (the 2019-10-04 rulings). The payment record names
// the creature (PaidCost.SacrificedObjects, ADR 0113 §1), and the damage
// is its power as it last existed on the battlefield (the 2024-04-12
// ruling, CR 608.2h): an anthem's bonus and its +1/+1 counters count,
// and a creature with negative power deals no damage (CR 107.1b). A copy
// deals the original's damage (CR 707.10).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "24227761-b50e-4b9e-93a2-e82d053b3e3d",
		Name:           "Fling",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		Targets:        TargetAny(),
		OnResolve:      damageEqualToSacrificedPower,
	})
}
