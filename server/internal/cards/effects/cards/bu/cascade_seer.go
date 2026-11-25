package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cascade Seer — Creature — Merfolk Wizard {3}{U}, 3/3 (Zendikar Rising,
// #5508):
//
//	"When this creature enters, scry X, where X is the number of creatures
//	 in your party. (Your party consists of up to one each of Cleric, Rogue,
//	 Warrior, and Wizard.)"
//
// X is the party (CR 700.8) as the trigger resolves, the Seer itself being
// the Wizard if nothing else is. Scry 0 does nothing (CR 701.22b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "08992e72-49b5-4d40-8b44-cb5abfe5ce0d",
		Name:         "Cascade Seer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cascade Seer — scry X, X the creatures in your party",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.Scry{N: effects.FinB9PartySize(ctx.Game, ctx.Controller())}.Apply(ctx)
				}))),
		},
	})
}
