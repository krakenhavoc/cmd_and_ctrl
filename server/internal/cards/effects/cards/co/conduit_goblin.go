package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conduit Goblin — Creature — Goblin Warrior {R}{W}, 2/2 (Modern Horizons
// 3):
//
//	"When this creature enters, you get {E}{E} (two energy counters).
//	 At the beginning of combat on your turn, you may pay {E}. If you do,
//	 another target creature you control gets +1/+0 and gains haste until
//	 end of turn."
//
// The entry is Aether Chaser's. The combat trigger is Additive
// Evolution's step trigger with a target chosen when it goes on the stack
// ("another": the Goblin itself is never legal), and the energy is paid
// as it resolves (ADR 0129 §3), as Aetherstorm Roc does. A target that is
// gone by then leaves the energy unpaid (CR 608.2b).
//
// No simplification.
func init() {
	combat := effects.On(game.EventStepBegan, effects.StepBegan(game.StepBeginCombat, true),
		"Conduit Goblin — you may pay {E}",
		effects.MayPayEnergyThen("Conduit Goblin", 1, "give another creature +1/+0 and haste", effects.GapB53PumpAndHaste))
	combat.TargetsFrom = effects.AnotherTarget(func(other effects.CardPredicate) *game.TargetSpec {
		return effects.TargetCreature("another target creature you control", effects.YouControl(), other)
	})
	effects.Register(effects.Spec{
		OracleID:     "0d3a38ff-4b75-4e97-9d8f-c54f441f8208",
		Name:         "Conduit Goblin",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Conduit Goblin", 2),
			combat,
		},
	})
}
