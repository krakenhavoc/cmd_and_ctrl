package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sneering Shadewriter — Creature — Vampire Warlock {4}{B}, 3/3 (Secrets of
// Strixhaven, #4456):
//
//	"Flying
//	 When this creature enters, each opponent loses 2 life and you gain 2
//	 life."
//
// Flying is a printed keyword; the enters trigger is the drain
// (EachOpponentLosesLife, then GainLife), a life change with no damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "21a094d6-de3d-48d4-a35d-b99617da55db",
		Name:            "Sneering Shadewriter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sneering Shadewriter — each opponent loses 2 life and you gain 2 life",
				Do(EachOpponentLosesLife{N: 2}, GainLife{Amount: 2})),
		},
	})
}
