package effects

// Cackling Imp — Creature — Imp {2}{B}{B}, 2/2 (Fifth Dawn, #5604):
//
//	"Flying
//	 {T}: Target player loses 1 life."
//
// The life loss is TargetPlayerLosesLife (Sweettooth Witch's body), a life
// change that names the Imp and no damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "adc8c55c-8fbe-4e07-b1be-8cb796db3d8a",
		Name:            "Cackling Imp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target player loses 1 life.",
			Cost:    TapCost(),
			Targets: TargetPlayer("target player"),
			Effect:  Do(TargetPlayerLosesLife{N: 1}),
		}},
	})
}
