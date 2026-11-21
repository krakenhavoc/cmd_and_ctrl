package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "49c7b7c6-c8dc-4871-9103-b057d2034c3e",
		Name:            "Vampire Sovereign",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Vampire Sovereign — target opponent loses 3 life and you gain 3 life",
				effects.AbilityBody(effects.TargetPlayerLosesLifeYouGain(3))),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
