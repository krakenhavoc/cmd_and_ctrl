package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hunted Lammasu — Creature — Lammasu {2}{W}{W}, 5/5 (Ravnica: City of
// Guilds, #5600):
//
//	"Flying
//	 When this creature enters, target opponent creates a 4/4 black Horror
//	 creature token."
//
// Hunted Horror's targeted enters trigger: the opponent creates the
// Horror, so they own and control it (CR 111.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a76667b5-2144-4395-a5c1-57cb3c29a99e",
		Name:            "Hunted Lammasu",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.RavB4TargetOpponentCreates("Hunted Lammasu — target opponent creates a 4/4 black Horror", "4/4 black Horror", 1),
		},
	})
}
