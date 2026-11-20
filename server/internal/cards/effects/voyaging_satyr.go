package effects

// Voyaging Satyr — Creature — Satyr Druid {1}{G}, 1/2 (Theros, #5561):
//
//	"{T}: Untap target land."
//
// Portent Tracker's untap row: any land, whoever controls it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bce18f34-848e-4dd2-9a38-6084c0695f78",
		Name:         "Voyaging Satyr",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Untap target land.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target land", Land()),
			Effect:  b22UntapEachLegalTarget,
		}},
	})
}
