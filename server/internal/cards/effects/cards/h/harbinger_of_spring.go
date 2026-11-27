package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harbinger of Spring — Creature — Spirit {4}{G}, 2/1 (Betrayers of Kamigawa, S82 OLD
// b596):
//
//	"Protection from non-Spirit creatures
//	 Soulshift 4"
//
// The protection is the closed quality game.ProtectionFromNonSubtypeCreatures (Spare from
// Evil's, #6253): it stops damage, enchanting and equipping, blocking and targeting by any
// creature that is not a Spirit, read as the source's creature types stand (CR 702.16).
// Soulshift 4 is the shared keyword trigger (CR 702.46a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "11c2e14c-cee5-480d-b1be-a6b7e245332a",
		Name:            "Harbinger of Spring",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.ProtectionFromNonSubtypeCreatures("Spirit")},
		Triggered:       []game.TriggeredAbility{effects.Soulshift(4, "Harbinger of Spring")},
	})
}
