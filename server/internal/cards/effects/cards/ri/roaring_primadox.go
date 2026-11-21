package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roaring Primadox — Creature — Beast {3}{G}, 4/4 (Magic 2013, #5568):
//
//	"At the beginning of your upkeep, return a creature you control to its
//	 owner's hand."
//
// Whitemane Lion's bounce on an upkeep trigger: mandatory, the controller
// picks as it resolves, and the Primadox itself is a legal pick.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a098296-6a64-48ba-af40-5b68bae06bec",
		Name:         "Roaring Primadox",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Roaring Primadox — return a creature you control", effects.Do(effects.ReturnOneYouControl{
				Match:    effects.MatchCreature,
				Question: "Roaring Primadox — return a creature you control to its owner's hand",
			})),
		},
	})
}
