package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Breaker of Armies — Creature — Eldrazi, {8}, 10/8:
//
//	"All creatures able to block this creature do so."
//
// #1684 leftover: Lure printed on the creature itself, the
// Prized Unicorn shape (AllAbleToBlockDoSo). No new machinery.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0be1ace3-2b3d-4465-b63a-be523f73df36",
		Name:         "Breaker of Armies",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{AllAbleToBlockDoSo()},
	})
}
