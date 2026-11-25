package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Underworld Coinsmith — Enchantment Creature — Human Cleric {W}{B}, 2/2
// (Commander 2015, S82 OLD #5546):
//
//	"Constellation — Whenever this creature or another enchantment you
//	 control enters, you gain 1 life.
//	 {W}{B}, Pay 1 life: Each opponent loses 1 life."
//
// Constellation is Whitewater Naiads' condition: its own entry (it is an
// enchantment) or another enchantment entering under your control. The
// activated drain pays {W}{B} and 1 life as costs, so it can't be activated
// at 0 life, and every opponent still in the game loses 1 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae06e0e5-8a42-4d1f-a629-94152e5f1fe6",
		Name:         "Underworld Coinsmith",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.ConstellationTriggers, "Underworld Coinsmith — you gain 1 life (constellation)",
				effects.Do(effects.GainLife{Amount: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{B}, Pay 1 life: Each opponent loses 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{W}{B}"), effects.PayLife(1)),
			Effect:  effects.Do(effects.EachOpponentLosesLife{N: 1}),
		}},
	})
}
