package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moonlight Geist — Creature — Spirit {2}{W}, 2/1:
//
//	"Flying
//	 {3}{W}: Prevent all combat damage that would be dealt to and dealt by
//	 this creature this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): one to-and-by record
// (Mod.AndDealtBy) pinned to this creature.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "12a8c546-d0d1-4c3d-bfca-c1e59cef349a",
		Name:            "Moonlight Geist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{effects.ToAndByThisRow(
			"{3}{W}: Prevent all combat damage that would be dealt to and dealt by this creature this turn.",
			effects.ManaCost("{3}{W}"))},
	})
}
