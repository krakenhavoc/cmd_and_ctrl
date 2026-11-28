package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Haunted Angel — Creature — Angel {2}{W}, 3/3 (Apocalypse, #5614):
//
//	"Flying
//	 When this creature dies, exile it and each other player creates a 3/3
//	 black Angel creature token with flying."
//
// The Angel is exiled from the graveyard if it is still there, and every
// player but its controller creates the token under their own control; the
// two instructions do not depend on each other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "437be0ad-2dd1-42c8-bf62-5af95228b2f8",
		Name:            "Haunted Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Haunted Angel — exile it; each other player creates a 3/3 black Angel creature token with flying",
				effects.S82OldB635HauntedAngel),
		},
	})
}
