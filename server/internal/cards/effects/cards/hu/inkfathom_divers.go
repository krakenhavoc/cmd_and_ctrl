package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inkfathom Divers — Creature — Merfolk Soldier {3}{U}{U}, 3/3 (Lorwyn):
//
//	"Islandwalk
//	 When this creature enters, look at the top four cards of your library,
//	 then put them back in any order."
//
// Sage Owl's look-at-the-top-four prompt on an enters trigger; islandwalk is
// the engine's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dbdf3c82-b0fc-479c-acce-f7e60be3a2b1",
		Name:            "Inkfathom Divers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Inkfathom Divers — look at the top four cards of your library, then put them back in any order", effects.Do(effects.LookAtTop{N: 4})),
		},
	})
}
