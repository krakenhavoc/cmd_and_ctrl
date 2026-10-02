package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raven Familiar — Creature — Bird, {2}{U}, 1/2:
//
//	"Flying
//	 Echo {2}{U} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, look at the top three cards of your library. Put one of them into your hand and the rest on the bottom of your library in any order."
//
// The look is Impulse's sentence (LookAtTopThenTakeOneRestOnBottom).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "47aabd9d-0f68-490a-a2ae-5cd9ef689be6",
		Name:            "Raven Familiar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Echo("Raven Familiar", "{2}{U}"),
			WhenThisEnters("Raven Familiar — look at the top three, take one",
				LookAtTopThenTakeOneRestOnBottom(3, "Raven Familiar — put one into your hand")),
		},
	})
}
