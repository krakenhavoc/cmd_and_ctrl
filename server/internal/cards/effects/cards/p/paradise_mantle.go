package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Paradise Mantle — Artifact — Equipment {0}:
//
//	"Equipped creature has "{T}: Add one mana of any color."
//	 Equip {1}"
//
// A granted mana ability that follows the Equipment (ADR 0093,
// GrantAbilitiesToAttached): moving the Mantle moves the ability. The
// equipped creature taps for it, so CR 302.6 still keeps a creature
// that entered this turn from doing so.
//
// No simplification.
const paradiseMantleGrant = "paradise-mantle/any-color"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "c1121b83-1ba2-473d-89c9-e3bbd4529072",
		Name:         "Paradise Mantle",
		Completeness: effects.CompletenessFull,
		Grants:       []effects.AbilityGrant{effects.AnyColorManaGrant(paradiseMantleGrant)},
		Static:       []game.StaticAbility{effects.GrantAbilitiesToAttached(paradiseMantleGrant)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
