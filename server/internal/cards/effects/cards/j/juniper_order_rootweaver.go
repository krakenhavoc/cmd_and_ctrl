package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Juniper Order Rootweaver — Creature — Human Druid {1}{W}, 2/2 (Dominaria
// United, #5487):
//
//	"Kicker {G} (You may pay an additional {G} as you cast this spell.)
//	 When this creature enters, if it was kicked, put a +1/+1 counter on
//	 target creature you control."
//
// The kicked intervening "if" (CR 603.4) gates the trigger, so an unkicked
// Rootweaver asks for no target; the target is chosen as it goes on the
// stack and may be the Rootweaver itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "5914d4d5-51b8-49c2-9459-bf5579df4990",
		Name:          "Juniper Order Rootweaver",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{G}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Juniper Order Rootweaver — kicked: put a +1/+1 counter on target creature you control",
				effects.Cmdr2S46PlusOneCounterOnClauseTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
