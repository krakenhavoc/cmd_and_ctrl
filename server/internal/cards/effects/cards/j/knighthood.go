package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knighthood — Enchantment {2}{W} (Urza's Legacy, #5623):
//
//	"Creatures you control have first strike."
//
// Fervor's anthem with first strike: a layer-6 grant (CR 613.1f) to every
// creature its controller controls, through TribalKeywordGrant with no
// tribe named.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22f008b3-c263-4dcc-855f-403da9ca4cd9",
		Name:         "Knighthood",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{YoursOnly: true}, "first strike"),
		},
	})
}
