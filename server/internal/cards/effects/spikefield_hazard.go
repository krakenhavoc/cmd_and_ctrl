package effects

// Spikefield Hazard // Spikefield Cave — modal double-faced card. This
// file is the FRONT face, Instant {R}:
//
//	"Spikefield Hazard deals 1 damage to any target. If a permanent dealt
//	 damage this way would die this turn, exile it instead."
//
// The back face, Spikefield Cave, is registered with the MDFC land
// cycle in mdfc_lands.go under "<oracle>#1"; the front keeps the bare
// oracle ID (Hagra Mauling's split).
//
// "A permanent dealt damage this way" is registered from the damage's
// continuation (ADR 0108 §1 decision 3), on any permanent that took more
// than 0 damage — a planeswalker or a battle as well as a creature.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "81036c9f-fe0a-45a7-bcd5-0d344f31055a",
		Name:         "Spikefield Hazard",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 1)),
		Targets:      TargetAny(),
		OnResolve:    damageAnyTargetExileIfDealtDies(fixedAmount(1), true),
	})
}
