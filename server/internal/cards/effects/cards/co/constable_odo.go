package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Constable Odo — Legendary Creature — Shapeshifter Detective
// {3}{W/U}{W/U}, 4/3 (Star Trek, #4973):
//
//	"When Constable Odo enters, choose one —
//	 • Tap target creature an opponent controls. Put two stun counters on
//	   it.
//	 • Investigate twice."
//
// A modal enters trigger (CR 700.2b): the mode and its target are chosen
// as the trigger goes on the stack. The first bullet taps its target and
// puts two stun counters on it (CR 122.1d); it does nothing if the target
// is no longer legal. Investigating twice creates two Clue tokens
// (CR 701.16a).
//
// No simplification.
func init() {
	odo := effects.WhenThisEnters("Constable Odo — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	odo.Modes = effects.ChooseOne(
		effects.ModeDoing("Tap target creature an opponent controls. Put two stun counters on it.",
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			effects.TrekB23TapAndTwoStunTheModesTarget),
		effects.ModeDoing("Investigate twice.", nil, func(_ *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.Investigate{N: 2}.Apply(ctx)
		}),
	)
	effects.Register(effects.Spec{
		OracleID:     "74b58f31-7634-4b51-a900-6fbfccfb3cd2",
		Name:         "Constable Odo",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{odo},
	})
}
