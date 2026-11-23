package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jett Reno, Problem Solver — Legendary Creature — Human Artificer
// {1}{U}, 1/2 (Star Trek, #4973):
//
//	"Whenever Jett Reno becomes tapped, surveil 1. If her power is 4 or
//	 greater, instead draw a card.
//	 Whenever you face a dilemma, put a +1/+1 counter on Jett Reno. (You
//	 face a dilemma as you choose one or more modes for a spell or
//	 ability.)"
//
// "Becomes tapped" is ThisBecameTapped over the tap and the attack
// events (#4534): a vigilance attack does not tap her. Her power is read
// as the trigger resolves (CR 608.2h), from her last-known information
// if she has left the battlefield, and decides which of the two
// instructions happens.
//
// "Face a dilemma" is the engine's EventFacedDilemma (#4973): the
// controller chose one or more modes as they cast a modal spell, put a
// modal activated ability on the stack, or answered a modal trigger's
// mode choice (CR 700.2a, 700.2b). A copy chooses nothing (CR 700.2g),
// and neither does a "choose up to" answered with no mode.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "492ba01e-c0fa-414c-95fd-df5b59212d2a",
		Name:         "Jett Reno, Problem Solver",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Jett Reno, Problem Solver — surveil 1, or draw a card if her power is 4 or greater",
				effects.Do(effects.If{Cond: jettRenoBig{}, Then: []effects.Applier{effects.DrawCards{N: 1}}},
					effects.If{Cond: jettRenoBig{Not: true}, Then: []effects.Applier{effects.Surveil{N: 1}}})),
			effects.On(game.EventFacedDilemma, effects.YouFacedADilemma,
				"Jett Reno, Problem Solver — put a +1/+1 counter on Jett Reno",
				effects.B35PutCounterOnSelf),
		},
	})
}

// jettRenoBig is the tap trigger's "if her power is 4 or greater" (Not:
// "otherwise"): her power now, or as she last existed on the battlefield
// (CR 608.2h). It reads the battlefield through the layers, so it
// declares a board read.
type jettRenoBig struct{ Not bool }

func (c jettRenoBig) Holds(ctx *effects.Context) bool {
	info, ok := ctx.SourcePermanent()
	return (ok && info.Power >= 4) != c.Not
}

func (jettRenoBig) Condition() (game.FootprintStep, bool) {
	return effects.CondStep(game.FootprintCondBoard, 0), true
}
