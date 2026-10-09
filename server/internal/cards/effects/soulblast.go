package effects

// Soulblast — Instant {3}{R}{R}{R}:
//
//	"As an additional cost to cast this spell, sacrifice all creatures
//	 you control.
//	 Soulblast deals damage to any target equal to the total power of
//	 the sacrificed creatures."
//
// The cost is #2097's SacrificeAllCost: every creature the caster
// controls as the spell is cast, chosen by nobody, gone before anyone
// can respond (CR 601.2h). A phased-out creature is not among them (CR
// 702.26b) and an indestructible one is (CR 701.21a). With no creatures
// the cost is paid with nothing and Soulblast deals no damage.
//
// The damage is ADR 0113 §1's record: each sacrificed creature's power
// as it last existed on the battlefield (CR 608.2h), summed with
// negatives included and the total floored at zero (CR 107.1b) —
// ctx.SacrificedTotalPower(), Corpse Cobble's reader. A creature that
// arrives after the cast is not counted, and a copy deals the
// original's amount (CR 707.10).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "18d4c57b-e2bf-47a0-8823-c4a79498a7ff",
		Name:           "Soulblast",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeAllCost("creatures you control", Creature()),
		Targets:        TargetAny(),
		OnResolve:      damageEqualToSacrificedTotalPower,
	})
}
