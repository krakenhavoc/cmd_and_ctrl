package effects

// Shacklegeist — Creature — Spirit {1}{U}, 2/2:
//
//	"Flying
//	 This creature can block only creatures with flying.
//	 Tap two untapped Spirits you control: Tap target creature you don't
//	 control."
//
// Air Bladder's block rule on itself, Catapult Squad's tap-two cost over
// Spirits.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7a1e4880-9abb-4018-84bd-91b5ec702c51",
		Name:            "Shacklegeist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules:      nemBlocksOnlyFlyers(OnSelf()),
		Activated: []ActivatedAbility{{
			Label:   "Tap two untapped Spirits you control: Tap target creature you don't control.",
			Cost:    tapTwoUntapped("two untapped Spirits you control", "Spirit"),
			Targets: TargetCreature("target creature you don't control", OpponentControls()),
			Effect:  b36TapChosenCreature,
		}},
	})
}
