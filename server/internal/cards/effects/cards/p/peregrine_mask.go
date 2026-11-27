package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Peregrine Mask — Artifact — Equipment {1} (Ravnica: City of Guilds, S82 OLD
// b582):
//
//	"Equipped creature has defender, flying, and first strike.
//	 Equip {2}"
//
// Bladed Pinions' keyword grant with three keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a742308c-4047-487f-b95b-0fa5bec2f09a",
		Name:         "Peregrine Mask",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.GrantToAttached("defender", "flying", "first strike")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
