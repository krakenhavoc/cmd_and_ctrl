package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Waterspout Djinn — Creature — Djinn {2}{U}{U}, 4/4 (Visions, S82 older sets
// batch 791, #5631):
//
//	"Flying
//	 At the beginning of your upkeep, sacrifice this creature unless you return
//	 an untapped Island you control to its owner's hand."
//
// Coral Atoll's payment moved to the upkeep: with no untapped Island the Djinn
// is sacrificed at once; otherwise the controller names one, or none, which
// sacrifices it. "Island" is the land type, so any untapped Island counts, and
// the word is read through the Djinn's text (CR 612.2). The Djinn stays only if
// the Island really reached a hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "050dac46-9ba0-4b8a-b61b-1c7ec6f3723a",
		Name:            "Waterspout Djinn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Old791WaterspoutDjinnReturn.Old791UpkeepTrigger("Waterspout Djinn — sacrifice this creature unless you return an untapped Island you control to its owner's hand"),
		},
	})
}
