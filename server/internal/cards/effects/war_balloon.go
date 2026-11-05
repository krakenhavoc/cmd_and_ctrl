package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// War Balloon — Artifact — Vehicle {2}{R}, 4/3 (Avatar: The Last
// Airbender, #4459):
//
//	"Flying
//	 {1}: Put a fire counter on this Vehicle.
//	 As long as this Vehicle has three or more fire counters on it, it's
//	 an artifact creature.
//	 Crew 3"
//
// The {1} ability has no timing limit and stacks fire counters on the
// Vehicle itself. The third line is a layer-4 static on the Vehicle
// (Phoenix Fleet Airship's shape): with three or more fire counters it is
// an artifact creature with its printed 4/3, crewed or not, and it stops
// being one if the counters are removed (unless it was crewed). Crew 3 is
// the shared ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "021b495d-236f-4274-988d-2dfd7103c9ef",
		Name:            "War Balloon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label: "As long as this Vehicle has three or more fire counters on it, it's an artifact creature.",
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && source.Counters["fire"] >= 3
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, t := range []string{"Artifact", "Creature"} {
					if !slices.Contains(c.Types, t) {
						c.Types = append(c.Types, t)
					}
				}
			},
		}},
		Activated: []ActivatedAbility{
			{
				Label:  "{1}: Put a fire counter on this Vehicle.",
				Cost:   ManaCost("{1}"),
				Effect: Do(CounterOnThis{Kind: "fire", N: 1}),
			},
			{
				Label:   "Crew 3",
				Cost:    CrewCost(3),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("War Balloon"),
			},
		},
	})
}
