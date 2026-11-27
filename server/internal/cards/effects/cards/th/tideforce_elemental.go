package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tideforce Elemental — Creature — Elemental {2}{U} 2/1 (Worldwake, S82 OLD
// batch 489, #5582):
//
//	"{U}, {T}: You may tap or untap another target creature.
//	 Landfall — Whenever a land you control enters, you may untap this
//	 creature."
//
// The tap-or-untap choice is TapOrUntapPermanent (Janjeet Sentry's), asked as
// the ability resolves; landfall is optional.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bda87ca2-91d0-48fa-bb82-a3396f0438ac",
		Name:         "Tideforce Elemental",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: You may tap or untap another target creature.",
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Targets: effects.Another(effects.TargetCreature("another target creature")),
			Effect:  effects.B38TapOrUntapTarget("Tideforce Elemental"),
		}},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Landfall("Tideforce Elemental — you may untap this creature", effects.B33UntapSelf),
				"Tideforce Elemental — untap it?"),
		},
	})
}
