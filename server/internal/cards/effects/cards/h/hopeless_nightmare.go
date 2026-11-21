package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hopeless Nightmare — Enchantment {B} (Wilds of Eldraine, #4471):
//
//	"When this enchantment enters, each opponent discards a card and
//	 loses 2 life.
//	 When this enchantment is put into a graveyard from the battlefield,
//	 scry 2.
//	 {2}{B}: Sacrifice this enchantment."
//
// The discard is Burglar Rat's: each opponent chooses a card from their
// own hand (CR 701.9a), and one with an empty hand discards nothing but
// still loses the life. The graveyard trigger and the sacrifice ability
// are Hopeful Vigil's: any trip from the battlefield to a graveyard
// scries 2, and the ability sacrifices on resolution, not as a cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e63b69c-d901-4a01-8dfa-88e044856239",
		Name:         "Hopeless Nightmare",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hopeless Nightmare — each opponent discards a card and loses 2 life",
				effects.Do(
					effects.Instruction(func(ctx *effects.Context) error { return effects.EachOpponentDiscardsOne(ctx.Game, ctx.Item) }),
					effects.EachOpponentLosesLife{N: 2},
				)),
			effects.On(game.EventLTB, effects.ThisDied, "Hopeless Nightmare — scry 2", effects.Do(effects.Scry{N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}: Sacrifice this enchantment.",
			Cost:    effects.ManaCost("{2}{B}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  effects.SacrificeDelayedTriggerSource,
		}},
	})
}
