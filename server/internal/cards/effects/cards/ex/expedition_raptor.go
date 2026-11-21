package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Expedition Raptor — Creature — Bird {3}{W}{W}, 2/2 (Foundations Jumpstart):
//
//	"Flying
//	 When this creature enters, support 2. (Put a +1/+1 counter on each of
//	 up to two other target creatures.)"
//
// Aerie Auxiliary's support (CR 701.41a): "other target creatures", so the
// Raptor cannot pick itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0a44da1d-7914-4592-a00f-7e40e0d07f91",
		Name:            "Expedition Raptor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersSupport("Expedition Raptor", 2, true),
		},
	})
}
