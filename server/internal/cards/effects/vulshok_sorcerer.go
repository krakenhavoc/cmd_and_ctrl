package effects

// Vulshok Sorcerer — Creature — Human Shaman Sorcerer {1}{R}{R}, 1/1 (Fifth Dawn, #5604):
//
//	"Haste
//	 {T}: This creature deals 1 damage to any target."
//
// Zuran Spellcaster's ping on a hasty body, so it can tap the turn it
// arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6f858405-dfe4-4ca2-b2e3-80ce8cda5522",
		Name:            "Vulshok Sorcerer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target.",
			Cost:    TapCost(),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  sourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
