package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Saprazzan Heir — Creature — Merfolk {1}{U}, 1/1 (Mercadian Masques, S82
// OLD, #5619):
//
//	"Whenever this creature becomes blocked, you may draw three cards."
//
// Chambered Nautilus's trigger for three cards: once per blocked attacker
// (CR 509.3c), the "may" asked as it resolves, and a player a draw limit stops
// from drawing three is not asked (#4016).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8685ebfd-d04d-47f5-a919-e118db1e2276",
		Name:         "Saprazzan Heir",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.On(game.EventBecomesBlocked, effects.Self,
				"Saprazzan Heir — you may draw three cards", effects.Do(effects.DrawCards{N: 3})),
				"Saprazzan Heir — draw three cards?", 3),
		},
	})
}
