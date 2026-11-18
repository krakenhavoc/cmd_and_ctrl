package effects

// Spectral Shepherd — Creature — Spirit {2}{W}, 2/2:
//
//	"Flying
//	 {1}{U}: Return target Spirit you control to its owner's hand."
//
// A targeted bounce of one of your own Spirits (itself included); a
// Spirit that left in response makes the ability do nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d28b7c07-ef65-4187-a15e-2180bdc593d1",
		Name:            "Spectral Shepherd",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}: Return target Spirit you control to its owner's hand.",
			Cost:    ManaCost("{1}{U}"),
			Targets: TargetCreature("target Spirit you control", OfCreatureType("Spirit"), YouControl()),
			Effect:  bounceChosenTarget,
		}},
	})
}
