package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// would_destroy_land.go — Equinox's granted ability (#6225):
//
//	"{T}: Counter target spell if it would destroy a land you control."
//
// The target clause is any spell; the condition is read as the ability
// resolves (game.SpellWouldDestroyALandOfForEffect), by playing the
// spell's resolution out on a copy of the game. So every catalog spell
// answers truthfully with nothing declared on it: a Stone Rain aimed at
// your land, an Armageddon, a Murder aimed at your animated land are
// countered; a Shock, a Stone Rain aimed at someone else's land, and one
// whose land target has already gone are not.
//
// Append-only: add a builder, never change what one means.

// CounterTargetSpellIfItWouldDestroyALandYouControl is the ability's
// body. "You" is the ability's controller.
func CounterTargetSpellIfItWouldDestroyALandYouControl(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	spell := targets[0].ID
	if !g.SpellWouldDestroyALandOfForEffect(spell, item.Controller) {
		return nil
	}
	return CounterTarget{StackID: spell}.Apply(ctx)
}
