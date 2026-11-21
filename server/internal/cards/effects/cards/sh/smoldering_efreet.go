package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smoldering Efreet — Creature — Efreet Monk {1}{R}, 2/2 (Fate Reforged,
// #5552):
//
//	"When this creature dies, it deals 2 damage to you."
//
// Callous Inspector's dies trigger: the Efreet deals the damage as it last
// existed on the battlefield (CR 608.2h) to its controller, so a prevention
// or a redirect sees it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "789b6a61-656e-40c0-be78-9c89c4ecfb0b",
		Name:         "Smoldering Efreet",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Smoldering Efreet — it deals 2 damage to you", effects.Do(effects.DamageController{N: 2})),
		},
	})
}
