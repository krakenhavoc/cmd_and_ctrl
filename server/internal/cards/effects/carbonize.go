package effects

// Carbonize — Instant {2}{R}:
//
//	Carbonize deals 3 damage to any target. If it's a creature, it can't be regenerated this turn, and if it would die this turn, exile it instead.
//
// Both riders are the spell's, not the damage's (the Disintegrate shape): a creature target is marked whether or not the damage is dealt (ADR 0108 §1, §2).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "a32795a2-a965-4a85-9944-fd9eed464e65",
		Name:         "Carbonize",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		OnResolve:    damageThenIfCreatureNoRegenExileIfDies(fixedAmount(3)),
	})
}
