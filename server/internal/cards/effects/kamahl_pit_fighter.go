package effects

// Kamahl, Pit Fighter — Legendary Creature — Human Barbarian {4}{R}{R},
// 6/1 (Odyssey, #5613):
//
//	"Haste (This creature can attack and {T} as soon as it comes under
//	 your control.)
//	 {T}: Kamahl deals 3 damage to any target."
//
// Zuran Spellcaster's pinger at 3. Haste is printed (CR 702.10), so the
// {T} row works the turn Kamahl arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "153bb46f-ee8a-4da3-aab4-ae884f6ff403",
		Name:            "Kamahl, Pit Fighter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Kamahl deals 3 damage to any target.",
			Cost:    TapCost(),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 3)),
			Effect:  sourceDealsDamageToEachLegalTarget(3),
		}},
	})
}
