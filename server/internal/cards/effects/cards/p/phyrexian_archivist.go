package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Archivist — Artifact Creature — Phyrexian Construct 4/5 {6} (March of the Machine, S82):
//
//	"Reach
//	 {2}, {T}: Put target card from a graveyard on the bottom of its owner's library."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cdef800d-033e-4095-964d-9b6066f9fd36",
		Name:            "Phyrexian Archivist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Put target card from a graveyard on the bottom of its owner's library.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
			Effect:  effects.SocB37PutTargetGraveyardCardOnBottom,
		}},
	})
}
