package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vampire Sovereign — Creature — Vampire Noble {3}{B}{B}, 3/4
// (Core Set 2019, #5527):
//
//	"Flying
//	When this creature enters, target opponent loses 3 life and you gain 3 life."
//
// Flying, and an enters trigger whose body is Sovereign's Bite's
// (targetPlayerLosesLifeYouGain) aimed at an opponent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "49c7b7c6-c8dc-4871-9103-b057d2034c3e",
		Name:            "Vampire Sovereign",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Vampire Sovereign — target opponent loses 3 life and you gain 3 life",
				abilityBody(targetPlayerLosesLifeYouGain(3))),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
