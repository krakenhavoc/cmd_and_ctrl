package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skulking Fugitive — Creature — Horror Mercenary {2}{B}, 3/4 (Mercadian Masques, S82 OLD batch 687, #5619):
//
//	"When this creature becomes the target of a spell or ability, sacrifice it."
//
// Skulking Ghost's watch on EventBecomesTarget narrowed to the Fugitive itself, open to a
// spell or ability of any controller. The trigger goes on the stack as the targets are chosen,
// so it is sacrificed before the targeting spell or ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4178ae56-2449-4617-aff8-f59d8fb96c66",
		Name:         "Skulking Fugitive",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.Self, "Skulking Fugitive — sacrifice it", effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
