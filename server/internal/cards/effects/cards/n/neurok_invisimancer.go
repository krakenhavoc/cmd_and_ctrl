package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Neurok Invisimancer — Creature — Human Wizard {1}{U}{U}, 2/1 (Scars of
// Mirrodin, #5577):
//
//	"This creature can't be blocked.
//	 When this creature enters, target creature can't be blocked this
//	 turn."
//
// Dimir Infiltrator's static and Clammy Prowler's "target creature can't
// be blocked this turn" (targetCantBeBlockedThisTurn) on a required
// enters trigger. The target may be the Invisimancer itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7425964-0028-4156-9037-b303b00e7e11",
		Name:         "Neurok Invisimancer",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Neurok Invisimancer — target creature can't be blocked this turn",
					effects.TargetCantBeBlockedThisTurn("Neurok Invisimancer")),
				effects.TargetCreature("target creature")),
		},
	})
}
