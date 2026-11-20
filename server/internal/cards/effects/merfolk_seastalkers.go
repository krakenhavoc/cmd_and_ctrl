package effects

// Merfolk Seastalkers — Creature — Merfolk Scout {3}{U}, 2/3 (Zendikar, #5583):
//
//	"Islandwalk (This creature can't be blocked as long as defending
//	 player controls an Island.)
//	 {2}{U}: Tap target creature without flying."
//
// Islandwalk is landwalk, an engine keyword (CR 702.14c); the tap is
// Cephalid Retainer's row at {2}{U}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "86d592e1-4b2d-4411-9370-48326cf55cf1",
		Name:            "Merfolk Seastalkers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}: Tap target creature without flying.",
			Cost:    ManaCost("{2}{U}"),
			Targets: TargetCreature("target creature without flying", Not(HasKeyword("flying"))),
			Effect:  tapChosenPermanent,
		}},
	})
}
