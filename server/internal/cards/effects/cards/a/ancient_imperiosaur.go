package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ancient Imperiosaur — Creature — Dinosaur {5}{G}{G} 6/6 (March of the
// Machine):
//
//	"Convoke
//	 Trample, ward {2}
//	 This creature enters with two +1/+1 counters on it for each creature
//	 that convoked it."
//
// Convoke is Spec.TapCost. The counters are an entry replacement
// (CR 614.1c) that reads the creatures the cast tapped
// (CountersPerConvoked), so they are on it as it enters. Ward is the
// ordinary ward trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3fe95820-cd7f-4f28-82bf-eb4350ea4179",
		Name:            "Ancient Imperiosaur",
		Completeness:    effects.CompletenessFull,
		TapCost:         effects.Convoke(),
		PrintedKeywords: []string{"trample"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			effects.CountersPerConvoked(game.CounterPlusOne, 2),
		},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Ancient Imperiosaur — ward {2}"),
		},
	})
}
