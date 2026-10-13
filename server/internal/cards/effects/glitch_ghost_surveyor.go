package effects

// Glitch Ghost Surveyor — Creature — Spirit Scout {2}{U}, 2/2:
//
//	"Flying
//	 Start your engines!
//	 Max speed — {3}, Exile this card from your graveyard: Draw a card."
//
// ADR 0138 (#2711). The graveyard ability is the shared Surveyor body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f0f83ab8-6185-4139-b41e-a11500036915",
		Name:            "Glitch Ghost Surveyor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", StartYourEngines},
		Activated:       []ActivatedAbility{maxSpeedSurveyorDraw()},
	})
}
