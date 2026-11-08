package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gargantuan Leech — Creature — Leech {7}{B}, 5/5 (The Lost Caverns of
// Ixalan, #4470):
//
//	"This spell costs {1} less to cast for each Cave you control and each
//	 Cave card in your graveyard.
//	 Lifelink"
//
// A self cost reduction (CR 601.2f) counted as the spell is cast, with
// Sinuous Benthisaur's count (caveCount): Cave lands you control plus
// Cave cards in your graveyard. The reduction spends against generic mana
// only, so the {B} is always paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "073f12d1-eaa1-4d08-bc1d-b216a3dc0c67",
		Name:            "Gargantuan Leech",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(func(q game.CostQuery) int {
				if q.Game == nil {
					return 0
				}
				return caveCount(q.Game, q.Controller, uuid.Nil)
			}, "This spell costs {1} less to cast for each Cave you control and each Cave card in your graveyard."),
		},
	})
}
