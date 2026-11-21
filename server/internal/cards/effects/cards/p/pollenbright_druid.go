package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pollenbright Druid — Creature — Elf Druid {1}{G}, 1/1 (Commander
// Masters):
//
//	"When this creature enters, choose one —
//	 • Put a +1/+1 counter on target creature.
//	 • Proliferate."
//
// An entry trigger with two modes (Adaptive Sporesinger's shape): the
// target is chosen only for the first, as the trigger goes on the stack
// (CR 700.2b), and the second is the proliferate primitive, which lets the
// controller choose any number of permanents and players with a counter.
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Pollenbright Druid — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = effects.ChooseOne(
		effects.ModeDoing("Put a +1/+1 counter on target creature.",
			effects.TargetCreature("target creature"), effects.PlusOneCounterOnTheModesTarget),
		effects.ModeSteps("Proliferate.", nil, effects.Proliferate{}),
	)
	effects.Register(effects.Spec{
		OracleID:     "828fe0c0-e280-4a8b-9e49-68cb34c2da97",
		Name:         "Pollenbright Druid",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
