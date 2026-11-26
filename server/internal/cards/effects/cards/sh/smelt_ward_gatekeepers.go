package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smelt-Ward Gatekeepers — Creature — Human Warrior {3}{R}, 2/4 (Dragon's
// Maze, S82 OLD b406, #5564):
//
//	"When this creature enters, if you control two or more Gates, gain control
//	 of target creature an opponent controls until end of turn. Untap that
//	 creature. It gains haste until end of turn."
//
// Ubul Sar Gatekeepers' intervening "if" (CR 603.4: checked as it triggers
// and again as it resolves) with Captivating Crew's threaten body. The target
// is chosen as the trigger goes on the stack (CR 603.3d); one that is no
// longer legal at resolution is left alone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c6d213b-448a-4231-9f6f-65c8ad49a5f1",
		Name:         "Smelt-Ward Gatekeepers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.InterveningIf(effects.CondControlTwoOrMoreOfSubtype("Gate"),
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ControlTwoOrMoreOfSubtype("Gate")),
					"Smelt-Ward Gatekeepers — if you control two or more Gates, gain control of target creature an opponent controls until end of turn, untap it, it gains haste",
					effects.S82OldB406ThreatenFirstTarget("Smelt-Ward Gatekeepers"))),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
