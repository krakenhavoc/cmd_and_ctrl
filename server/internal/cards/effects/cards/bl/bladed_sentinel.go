package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bladed Sentinel — Artifact Creature — Construct {4}, 2/4 (Mirrodin Besieged,
// #5576):
//
//	"{W}: This creature gains vigilance until end of turn."
//
// Wall of Tanglecord's self grant (somB1SelfGainsUntilEOT): a layer-6
// vigilance on the Sentinel that ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00f579f4-2d18-4c7c-91b2-34726b384535",
		Name:         "Bladed Sentinel",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.SomB1SelfGainsUntilEOT("Bladed Sentinel", "{W}", "vigilance")},
	})
}
