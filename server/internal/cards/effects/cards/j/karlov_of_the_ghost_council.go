package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Karlov of the Ghost Council — Legendary Creature — Spirit Advisor {W}{B},
// 2/2 (Commander 2015, #5546):
//
//	"Whenever you gain life, put two +1/+1 counters on Karlov.
//	 {W}{B}, Remove six +1/+1 counters from Karlov: Exile target creature."
//
// The life-gain trigger is Ajani's Pridemate's (once per life-gain event)
// with two counters. The activated ability's cost removes six +1/+1
// counters from Karlov itself (RemoveCountersFromThis), so it can't be
// activated with fewer; the exile is Hanged Executioner's
// exileFirstLegalCardTarget.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "647d7988-9a5d-4700-9238-938d0d5fae12",
		Name:         "Karlov of the Ghost Council",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LifeGainPaying(
				effects.WheneverYouGainLife("Karlov of the Ghost Council — put two +1/+1 counters on it",
					effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
				game.LifeGainPayoff{Counters: 2}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{B}, Remove six +1/+1 counters from Karlov: Exile target creature.",
			Cost:    effects.Plus(effects.ManaCost("{W}{B}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 6)),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.ExileFirstLegalCardTarget,
		}},
	})
}
