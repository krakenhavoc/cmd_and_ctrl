package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Edgewall Pack — Creature — Dog {3}{R}, 3/3 (Wilds of Eldraine, #4471):
//
//	"Menace (This creature can't be blocked except by two or more
//	 creatures.)
//	 When this creature enters, create a 1/1 black Rat creature token with
//	 "This token can't block.""
//
// Menace is the engine's keyword (CR 702.111b). The entry trigger makes
// Twisted Sewer-Witch's Rat, the catalog token whose own text is "This
// token can't block", under the trigger's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ef702a1a-4ace-43e0-8e86-dcceb0b1e4d5",
		Name:            "Edgewall Pack",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Edgewall Pack — create a 1/1 black Rat creature token with \"This token can't block.\"",
				effects.Do(effects.CreateToken{Template: effects.CantBlockRatToken(), N: 1})),
		},
	})
}
