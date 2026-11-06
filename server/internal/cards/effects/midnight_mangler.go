package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Midnight Mangler — Artifact — Vehicle {1}{U}, 3/3 (Aetherdrift, #4464):
//
//	"During turns other than yours, this Vehicle is an artifact creature.
//	 Crew 2 (Tap any number of creatures you control with total power 2
//	 or more: This Vehicle becomes an artifact creature until end of
//	 turn.)"
//
// War Balloon's conditional animation (selfIsArtifactCreatureWhile), a
// layer-4 static (CR 613.1d) read live, so it is a creature on every
// opponent's turn and stops being one when its controller's turn begins
// (CR 611.3a). It has its printed 3/3 whenever it is a creature
// (CR 301.7b). Crew is the engine's (CR 702.122a), for its own turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c9f8651a-4118-4d80-902e-60473e81d3a9",
		Name:         "Midnight Mangler",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{selfIsArtifactCreatureWhile(
			"During turns other than yours, this Vehicle is an artifact creature.",
			func(g *game.Game, source *game.Card) bool { return !isActivePlayer(g, source.Controller) },
		)},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Midnight Mangler"),
		}},
	})
}
