package effects

// Goblin Surveyor — Creature — Goblin Scout {2}{R}, 3/2:
//
//	"Trample
//	 Start your engines!
//	 Max speed — {3}, Exile this card from your graveyard: Draw a card."
//
// ADR 0138 (#2122). The ability functions from the graveyard (CR 113.6,
// ADR 0062), and so does the max speed ability that grants it
// (CR 702.178b), so it works there while its owner's speed is 4. The
// speed comes from any permanent with start your engines!, this one
// included while it was on the battlefield: nothing lowers a speed. The
// body is the shared Surveyor one (speed_surveyor.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "237758f3-84b2-4e5b-a2c7-b28c893e2790",
		Name:            "Goblin Surveyor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", StartYourEngines},
		Activated:       []ActivatedAbility{maxSpeedSurveyorDraw()},
	})
}
