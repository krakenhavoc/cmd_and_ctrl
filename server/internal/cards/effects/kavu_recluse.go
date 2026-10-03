package effects

// Kavu Recluse — Creature — Kavu {2}{R}, 2/2:
//
//	"{T}: Target land becomes a Forest until end of turn."
//
// ADR 0109 §1 (#1881): CR 305.7 from a resolved ability. Until end of
// turn the land's land types are replaced by Forest (its other subtypes
// stay, CR 205.1a), it loses the abilities its rules text gives it, keeps
// any another effect granted it, and taps for {G} (CR 305.6). A land
// that has become an illegal target by resolution is left alone (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d44ff432-7485-448f-a08d-8417e18b8820",
		Name:         "Kavu Recluse",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target land becomes a Forest until end of turn.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target land", Land()),
			Effect:  TargetLandBecomesUntilEOT("Kavu Recluse", "Forest"),
		}},
	})
}
