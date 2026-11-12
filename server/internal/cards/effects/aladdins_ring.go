package effects

// Aladdin's Ring — Artifact {8} (Arabian Nights, #5641):
//
//	"{8}, {T}: This artifact deals 4 damage to any target."
//
// The Ring is the source of the damage, as it last existed if it left
// in response.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7810915f-df4d-445a-8ba8-fe6ed3abf1ae",
		Name:         "Aladdin's Ring",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{8}, {T}: This artifact deals 4 damage to any target.",
			Purpose: ForTargets(DamageToTarget(0, 4)),
			Cost:    Plus(ManaCost("{8}"), TapCost()),
			Targets: TargetAny(),
			Effect:  fourDamageToFirstLegalTarget,
		}},
	})
}
