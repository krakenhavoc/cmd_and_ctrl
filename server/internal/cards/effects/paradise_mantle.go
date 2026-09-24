package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c1121b83-1ba2-473d-89c9-e3bbd4529072",
		Name:         "Paradise Mantle",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{AnyColorManaGrant(paradiseMantleGrant)},
		Static:       []game.StaticAbility{GrantAbilitiesToAttached(paradiseMantleGrant)},
		Activated:    []ActivatedAbility{EquipAbility("{1}")},
	})
}
