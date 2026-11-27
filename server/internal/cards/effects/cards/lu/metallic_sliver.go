package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Metallic Sliver — {1} Artifact Creature — Sliver, 1/1 (Tempest, S82 OLD batch
// 770, #5628): no abilities. The card is what its type line says.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2dfafd63-83fb-4783-8e12-30018680a905",
		Name:         "Metallic Sliver",
		Completeness: effects.CompletenessFull,
	})
}
