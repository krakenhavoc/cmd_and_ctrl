package effects

// Muraganda Raceway — Land:
//
//	"Start your engines!
//	 {T}: Add {C}.
//	 Max speed — {T}: Add {C}{C}."
//
// ADR 0136 (#2122). Two mana abilities; the second can be activated
// only while its controller has max speed (CR 702.178a), and the
// auto-tapper plans it only then (it reads the ability's Condition).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b5fa5651-d714-44d6-867b-be0e3224b7ed",
		Name:            "Muraganda Raceway",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		ManaAbilities: []ManaAbility{
			{Cost: ManaAbilityCost{Tap: true}, Produced: "{C}", Label: "{T}: Add {C}."},
			MaxSpeedMana(ManaAbility{Cost: ManaAbilityCost{Tap: true}, Produced: "{C}{C}", Label: "Max speed — {T}: Add {C}{C}."}),
		},
	})
}
