package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silumgar's Scorn — Instant {U}{U} (Dragons of Tarkir):
//
//	"As an additional cost to cast this spell, you may reveal a Dragon card from
//	 your hand.
//	 Counter target spell unless its controller pays {1}. If you revealed a Dragon
//	 card or controlled a Dragon as you cast this spell, counter that spell
//	 instead."
//
// The reveal is the optional behold cost and "controlled a Dragon as you cast it"
// a value recorded at announcement (ControlsADragonAsCast); either one turns the
// soft counter into a hard one.
//
// No simplification.
func init() {
	const dragonKey = "controlled_a_dragon"
	effects.Register(effects.Spec{
		OracleID:      "1fb3b5a3-ca4b-4cbf-ab58-71c960f3efd0",
		Name:          "Silumgar's Scorn",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove},
		OptionalCosts: []game.AdditionalCost{effects.OptionalBehold("a", "Dragon")},
		CastTimeValues: []game.CastTimeValue{{
			Key:   dragonKey,
			Value: effects.ControlsADragonAsCast,
		}},
		Targets: effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			if controlled, _ := ctx.CastValue(dragonKey); ctx.WasBeheld() || controlled > 0 {
				return effects.CounterTarget{StackID: targets[0].ID}.Apply(ctx)
			}
			return effects.CounterUnlessPaid{
				StackID:  targets[0].ID,
				Cost:     "{1}",
				Question: "Silumgar's Scorn — pay {1} or your spell is countered",
			}.Apply(ctx)
		},
	})
}
