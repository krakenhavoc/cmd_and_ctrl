package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lumbering Satyr — Creature — Satyr Beast {2}{G}{G}, 5/4 (Mercadian Masques, S82 OLD batch 688, #5619):
//
//	"All creatures have forestwalk.
//	(They can't be blocked as long as defending player controls a Forest.)"
//
// A layer-6 keyword grant (CR 613.1f) of the engine's own forestwalk to every creature at the table, an opponent's and its own included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fba1329d-d8bf-4c5e-9cf1-5410a42f732e",
		Name:         "Lumbering Satyr",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsCreature()
			}, "forestwalk"),
		},
	})
}
