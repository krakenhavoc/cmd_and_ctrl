package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wedgelight Rammer — Artifact — Spacecraft {3}{W}, 3/4 (Edge of
// Eternities, #4461):
//
//	"When this Spacecraft enters, create a 2/2 colorless Robot artifact
//	 creature token.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 9+.)
//	 9+ | Flying, first strike"
//
// Pinnacle Kill-Ship's station lines (Station, SpacecraftAt and
// ThresholdKeywords, CR 702.184a and 721.2a) with Mechan Assembler's
// Robot made by an entry trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "03259ab0-caa8-4620-b070-127e2c712252",
		Name:         "Wedgelight Rammer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Wedgelight Rammer — create a 2/2 colorless Robot artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 colorless Robot artifact"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{effects.Station()},
		Static: append(
			effects.SpacecraftAt(9, 3, 4),
			effects.ThresholdKeywords(9, "flying", "first strike"),
		),
	})
}
