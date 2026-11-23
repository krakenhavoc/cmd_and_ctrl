package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Breaker of Armies — Creature — Eldrazi, {8}, 10/8:
//
//	"All creatures able to block this creature do so."
//
// #1684 leftover: Lure printed on the creature itself, the
// Prized Unicorn shape (AllAbleToBlockDoSo). No new machinery.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0be1ace3-2b3d-4465-b63a-be523f73df36",
		Name:         "Breaker of Armies",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AllAbleToBlockDoSo()},
	})
}
