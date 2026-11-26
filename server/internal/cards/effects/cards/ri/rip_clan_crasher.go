package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rip-Clan Crasher — {R}{G} Creature — Human Warrior 2/2 (Shards of Alara, S82 OLD b515):
//
//	"Haste"
//
// Keywords only: each is read through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5d45fa1f-110c-4bc8-a0fc-ed04556cf639",
		Name:            "Rip-Clan Crasher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
