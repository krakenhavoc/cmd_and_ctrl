package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daggerfang Duo — Creature — Rat Squirrel {2}{B}, 3/2 (Bloomburrow,
// #4467):
//
//	"Deathtouch
//	 When this creature enters, you may mill two cards."
//
// Deathcap Marionette's text on a bigger body. Deathtouch is the
// printed keyword (CR 702.2a); the enters trigger is optional, asked as
// it resolves, and mills its controller's top two cards (CR 701.17a).
// Declining mills nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "43ccc568-2509-49b6-99d7-c56900d1565a",
		Name:            "Daggerfang Duo",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisEnters("Daggerfang Duo — mill two cards", Do(MillCards{N: 2})),
				"Daggerfang Duo — mill two cards?"),
		},
	})
}
