package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Foul Familiar — Creature — Spirit {2}{B}, 3/1 (Ice Age, S82 OLD batch 851,
// #5636):
//
//	"This creature can't block.
//	 {B}, Pay 1 life: Return this creature to its owner's hand."
//
// "Can't block" is Aesthir Glider's self restriction. The return is Brightling's
// (the ability's effect, not a cost, so it can be answered; a Familiar that left
// and came back is a new object and stays, CR 400.7) behind a {B} plus 1 life
// cost paid as the ability is activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "90767f52-fa6d-4ef3-ae54-d8e3881971c1",
		Name:         "Foul Familiar",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Pay 1 life: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.PayLife(1)),
			Effect:  effects.ReturnThisPermanentToOwnersHand,
		}},
	})
}
