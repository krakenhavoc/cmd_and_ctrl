package effects

// Vengeful Devil — Creature — Devil {1}{R}, 1/1 (Modern Horizons, #5519):
//
//	"Haste
//	 Morbid — {T}: This creature deals 1 damage to any target. Activate only
//	 if a creature died this turn."
//
// Anaba Shaman's ping for just {T}, behind the shared morbid activation
// condition (any creature, anyone's, this turn).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ba771caa-2cfb-4373-bbd9-6160b0543010",
		Name:            "Vengeful Devil",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:     "Morbid — {T}: This creature deals 1 damage to any target. Activate only if a creature died this turn.",
			Purpose:   ForTargets(DamageToTarget(0, 1)),
			Cost:      TapCost(),
			Targets:   TargetAny(),
			Condition: ACreatureDiedThisTurn(),
			Effect:    DealDamageToTheTarget(1),
		}},
	})
}
