package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rimewind Taskmage — Creature — Human Wizard {1}{U}, 1/2 (Coldsnap):
//
//	"{1}, {T}: You may tap or untap target permanent. Activate only if you
//	 control four or more snow permanents."
//
// The gate is read on activation only (it is not an intervening "if"). "You
// may" is covered by the choice itself: tapping a tapped permanent and
// untapping an untapped one do nothing, so the controller can always decline
// by picking the idle action.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5d9248e-9cff-4f80-b6ed-7d368099e97b",
		Name:         "Rimewind Taskmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{1}, {T}: You may tap or untap target permanent. Activate only if you control four or more snow permanents.",
			Purpose:   game.Purpose{Answers: game.AnswerRestrict},
			Cost:      effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Condition: effects.ControlsAtLeast(4, func(c game.Card) bool { return c.HasSupertype("snow") }),
			Targets:   effects.TargetPermanent("target permanent"),
			Effect:    effects.B38TapOrUntapTarget("Rimewind Taskmage"),
		}},
	})
}
