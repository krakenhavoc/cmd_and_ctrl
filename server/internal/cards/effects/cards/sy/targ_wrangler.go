package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Targ Wrangler — Creature — Klingon Warrior {2}{R}, 5/3 (Star Trek,
// #4973):
//
//	"Menace (This creature can't be blocked except by two or more
//	 creatures.)
//	 At the beginning of your end step, if this creature attacked this
//	 turn, sacrifice a nonland permanent."
//
// Menace is the printed keyword. The "if" is an intervening if (CR
// 603.4): the ability triggers only if this object attacked this turn,
// and checks again as it resolves, reading the object it came from (its
// last known information once it has left, CR 608.2h), so a Wrangler
// that attacked and then left still makes its controller sacrifice. The
// sacrifice is not targeted: as it resolves its controller chooses any
// nonland permanent they control, the Wrangler itself included, and with
// none nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a066eeb1-61fc-448b-bbab-650d672652ad",
		Name:            "Targ Wrangler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.TrekB18CondThisAttackedThisTurn{},
				effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, effects.TrekB18ThisAttackedThisTurn),
					"Targ Wrangler — sacrifice a nonland permanent",
					effects.Do(effects.Instruction(func(ctx *effects.Context) error {
						ctx.Game.PlayerSacrificesForEffect(ctx.Source(), ctx.Controller(),
							effects.SacrificeSpec("a nonland permanent", effects.Nonland()),
							"Targ Wrangler — sacrifice a nonland permanent")
						return nil
					})))),
		},
	})
}
