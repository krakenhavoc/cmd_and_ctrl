package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwarfhold Champion — {1}{W} Creature — Dwarf Warrior 3/1 (Adventures in
// the Forgotten Realms, #5499):
//
//	"As long as this creature is equipped, it gets +0/+2."
//
// Armed Assailant's "as long as this creature is equipped" gate
// (isEquippedCreature, read on every layer recompute) on a layer 7c pump.
//
// No simplification.
func init() {
	whileEquipped := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.SelfOnly(target, g, source) && effects.IsEquippedCreature(g, *target)
	}
	pump := effects.B16Anthem(whileEquipped, 0, 2)
	pump.Label = "As long as this creature is equipped, it gets +0/+2."
	effects.Register(effects.Spec{
		OracleID:     "3db19689-73bb-48b2-a464-63615eb6bba0",
		Name:         "Dwarfhold Champion",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{pump},
	})
}
