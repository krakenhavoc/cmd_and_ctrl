package effects

// Combust — Instant {1}{R}:
//
//	"This spell can't be countered.
//	 Combust deals 5 damage to target white or blue creature. The damage
//	 can't be prevented."
//
// Both riders are the spell's own: CantBeCountered, and its damage's
// "can't be prevented" (Spec.SpellDamageCantBePrevented, ADR 0107 §5),
// which the stack chip shows and the prevention gate reads as the
// damage is dealt.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:                   "f1ba56e3-3d09-4610-88e7-8681b5736105",
		Name:                       "Combust",
		Completeness:               CompletenessFull,
		CantBeCountered:            true,
		SpellDamageCantBePrevented: Always(),
		Targets:                    TargetCreature("target white or blue creature", Or(OfColor("W"), OfColor("U"))),
		Purpose:                    ForTargets(DamageToTarget(0, 5)),
		OnResolve:                  damageToFirstTarget(5),
	})
}
