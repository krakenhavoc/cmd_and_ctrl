package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vault Skirge — Artifact Creature — Phyrexian Imp {1}{B/P}, 1/1 (New Phyrexia, #5575):
//
//	"({B/P} can be paid with either {B} or 2 life.)
//	 Flying
//	 Lifelink (Damage dealt by this creature also causes you to gain that much life.)"
//
// The Phyrexian symbol is paid by the engine at cast (CR 107.4f); both
// keywords are the engine's (CR 702.9, CR 702.15).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "64cf5d59-7bcd-4b0b-a160-c8468d4c0f60",
		Name:            "Vault Skirge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
	})
}
