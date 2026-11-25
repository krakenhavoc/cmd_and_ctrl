package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight of the Ebon Legion — Creature — Vampire Knight {B}, 1/2 (Core Set 2020,
// #5518):
//
//	"{2}{B}: This creature gets +3/+3 and gains deathtouch until end of turn.
//	 At the beginning of your end step, if a player lost 4 or more life this
//	 turn, put a +1/+1 counter on this creature. (Damage causes loss of life.)"
//
// The trigger is Y'shtola's intervening "if" over every seated player's life
// lost this turn, the controller's own losses included, checked as the end step
// begins and again as it resolves (CR 603.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b4eaa50-dff6-4825-97da-42b2f5554218",
		Name:         "Knight of the Ebon Legion",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}: This creature gets +3/+3 and gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{2}{B}"),
			Effect:  effects.OldB156KnightPump,
		}},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(
				effects.CondBoard{
					Name: "a player lost 4 or more life this turn",
					Test: func(ctx *effects.Context) bool { return effects.B24AnyPlayerLostAtLeastThisTurn(ctx.Game, 4) },
				},
				effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return effects.ByYou(ev, source, lki, g) && effects.B24AnyPlayerLostAtLeastThisTurn(g, 4)
				}, "Knight of the Ebon Legion — put a +1/+1 counter on it",
					effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}
