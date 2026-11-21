package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "c9f8651a-4118-4d80-902e-60473e81d3a9",
		Name:         "Midnight Mangler",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{effects.SelfIsArtifactCreatureWhile(
			"During turns other than yours, this Vehicle is an artifact creature.",
			func(g *game.Game, source *game.Card) bool { return !effects.IsActivePlayer(g, source.Controller) },
		)},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Midnight Mangler"),
		}},
	})
}
