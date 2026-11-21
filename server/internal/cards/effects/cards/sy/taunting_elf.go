package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Taunting Elf — Creature — Elf, {G}, 0/1:
//
//	"All creatures able to block this creature do so."
//
// Prized Unicorn's Lure on a one-drop (#1684) — the chump that clears
// the way for everything attacking beside it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aef07e62-ac7f-4a5b-9732-136052d88a06",
		Name:         "Taunting Elf",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AllAbleToBlockDoSo()},
	})
}
