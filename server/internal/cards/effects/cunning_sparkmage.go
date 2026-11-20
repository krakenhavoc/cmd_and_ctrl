package effects

// Cunning Sparkmage — Creature — Human Shaman {2}{R}, 0/1 (Worldwake, #5582):
//
//	"Haste
//	 {T}: This creature deals 1 damage to any target."
//
// Anaba Shaman's ping without the mana, on a creature with haste so it
// can tap the turn it arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ab4a25d8-e1d0-454e-9d06-f7960a0352bc",
		Name:            "Cunning Sparkmage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target.",
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Cost:    TapCost(),
			Targets: TargetAny(),
			Effect:  DealDamageToTheTarget(1),
		}},
	})
}
