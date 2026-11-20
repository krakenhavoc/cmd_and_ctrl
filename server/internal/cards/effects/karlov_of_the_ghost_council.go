package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "647d7988-9a5d-4700-9238-938d0d5fae12",
		Name:         "Karlov of the Ghost Council",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			LifeGainPaying(
				WheneverYouGainLife("Karlov of the Ghost Council — put two +1/+1 counters on it",
					Do(CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
				game.LifeGainPayoff{Counters: 2}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{W}{B}, Remove six +1/+1 counters from Karlov: Exile target creature.",
			Cost:    Plus(ManaCost("{W}{B}"), RemoveCountersFromThis(game.CounterPlusOne, 6)),
			Targets: TargetCreature("target creature"),
			Effect:  exileFirstLegalCardTarget,
		}},
	})
}
