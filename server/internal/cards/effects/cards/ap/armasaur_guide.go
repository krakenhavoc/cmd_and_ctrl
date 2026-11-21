package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armasaur Guide — Creature — Dinosaur {4}{W}, 4/4 (Foundations, #4465):
//
//	"Vigilance
//	 Whenever you attack with three or more creatures, put a +1/+1
//	 counter on target creature you control."
//
// "Attack with three or more creatures" is Meddling Youths' count: the
// EventAttack events of one declaration by you, settled as a batch, one
// trigger when there are at least three (ThatManyEventsAtLeast). The Guide
// need not be among the attackers, and an additional combat's declaration
// is its own batch. The target is chosen as the trigger goes on the stack;
// any creature you control will do, attacking or not, the Guide included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1d4e7422-2e26-49ce-beef-08f61583089a",
		Name:            "Armasaur Guide",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WithAtBatchEnd(effects.Targeting(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source != nil && effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Armasaur Guide — put a +1/+1 counter on target creature you control (you attacked with three or more creatures)",
				effects.PlusOneCounterOnTheTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
				effects.ThatManyEventsAtLeast(3)),
		},
	})
}
