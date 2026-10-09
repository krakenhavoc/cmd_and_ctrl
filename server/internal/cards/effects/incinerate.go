package effects

// Incinerate — Instant {1}{R}:
//
//	"Incinerate deals 3 damage to any target. A creature dealt damage
//	 this way can't be regenerated this turn."
//
// "A creature dealt damage this way" is read from the damage's
// continuation (ADR 0108 §2): a creature whose damage was all prevented
// can still be regenerated.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "d8fd7a34-8418-4e98-b79b-119c4348c667",
		Name:         "Incinerate",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		OnResolve:    damageAnyTargetNoRegenIfDealt(fixedAmount(3)),
	})
}
