package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gate Smasher — Artifact — Equipment {3} (Dragons of Tarkir):
//
//	"This Equipment can be attached only to a creature with toughness 4 or greater.
//	 Equipped creature gets +3/+0 and has trample.
//	 Equip {3}"
//
// The attach restriction is checked attached or not (CR 301.5b): a host that
// drops below toughness 4 unattaches it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3425598f-d51c-4f92-8fda-42e96e4c537a",
		Name:         "Gate Smasher",
		Completeness: effects.CompletenessFull,
		AttachOnlyTo: effects.AttachedOnlyTo("a creature with toughness 4 or greater", effects.ToughnessGE(4)),
		Static:       []game.StaticAbility{effects.PumpAttached(3, 0), effects.GrantToAttached("trample")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
