package effects

// Touchstone — Artifact {2} (Weatherlight, #5629):
//
//	"{T}: Tap target artifact you don't control."
//
// Auriok Transfixer's tap with no mana and an artifact another player
// controls.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92382bdb-84cf-4456-8ca1-8a19badae7ad",
		Name:         "Touchstone",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Tap target artifact you don't control.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target artifact you don't control", Artifact(), OpponentControls()),
			Effect:  tapChosenPermanent,
		}},
	})
}
