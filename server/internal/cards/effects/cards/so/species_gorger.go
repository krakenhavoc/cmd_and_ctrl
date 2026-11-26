package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Species Gorger — Creature — Frog Beast {3}{G}{U}, 6/6 (Dragon's Maze, S82
// OLD b405, #5564):
//
//	"At the beginning of your upkeep, return a creature you control to its
//	 owner's hand."
//
// Roaring Primadox's trigger: mandatory, the controller picks as it resolves
// (not a target), and the Gorger itself is a legal pick.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e741c79-5ea4-457d-a7d3-b11154459246",
		Name:         "Species Gorger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Species Gorger — return a creature you control", effects.Do(effects.ReturnOneYouControl{
				Match:    effects.MatchCreature,
				Question: "Species Gorger — return a creature you control to its owner's hand",
			})),
		},
	})
}
