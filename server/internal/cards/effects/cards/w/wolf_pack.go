package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wolf Pack — Creature — Wolf {6}{G}{G}, 7/6
// (Portal Three Kingdoms, #5622):
//
//	"You may have this creature assign its combat damage as though it weren't blocked."
//
// The static is Lone Wolf's: a layer 6 grant of the creature's own flag (CR
// 510.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec58a82f-8c78-4f88-90d5-e32d83be7a48",
		Name:         "Wolf Pack",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.S82OldB614AssignAsThoughUnblocked()},
	})
}
