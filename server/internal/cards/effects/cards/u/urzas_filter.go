package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Urza's Filter — Artifact {4} (Invasion):
//
//	"Multicolored spells cost {2} less to cast."
//
// Arcane Melee's reduction (generic only, every player's spells) over spells
// of two or more colours, read through the spell's effective colours (CR
// 105.2, CR 702.114a devoid is colourless).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46484242-28ed-43d2-b33a-1e5c1771acca",
		Name:         "Urza's Filter",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(2, "Multicolored spells cost {2} less to cast.", func(q game.CostQuery) bool {
				return len(q.Card.EffectiveColors()) >= 2
			}),
		},
	})
}
