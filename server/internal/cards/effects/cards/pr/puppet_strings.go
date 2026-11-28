package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Puppet Strings — Artifact {3} (Tempest, S82 OLD b754, #5628):
//
//	"{2}, {T}: You may tap or untap target creature."
//
// Rimewind Taskmage's tap-or-untap choice on a creature: the controller picks
// tap or untap as the ability resolves, and the idle action is the way to
// decline.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb810d5c-903e-4ab4-874c-e49a686ca367",
		Name:         "Puppet Strings",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: You may tap or untap target creature.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.B38TapOrUntapTarget("Puppet Strings"),
		}},
	})
}
