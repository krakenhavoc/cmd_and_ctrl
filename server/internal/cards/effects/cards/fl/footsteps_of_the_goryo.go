package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Footsteps of the Goryo — Sorcery — Arcane {2}{B} (Saviors of Kamigawa,
// #5601):
//
//	"Return target creature card from your graveyard to the battlefield.
//	 Sacrifice that creature at the beginning of the next end step."
//
// Balduvian Atrocity's return without the haste: the creature comes back under
// your control, and a delayed trigger (CR 603.7) sacrifices that object at the
// next end step, only while it is still the permanent that returned (CR 400.7).
// A target gone by resolution makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0e8d9f0-86e9-46b7-ba6f-2976ff9148d6",
		Name:         "Footsteps of the Goryo",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
		OnResolve:    effects.OldB587FootstepsReturn,
	})
}
