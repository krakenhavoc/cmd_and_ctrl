package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shapers of Nature — Creature — Merfolk Shaman {1}{G}{U}, 3/3 (Ixalan,
// #5532):
//
//	"{3}{G}: Put a +1/+1 counter on target creature.
//	 {2}{U}, Remove a +1/+1 counter from a creature you control: Draw a
//	 card."
//
// The first ability is Ghave's counter body behind a mana cost. The
// second takes its counter off a creature chosen as the cost is paid
// (Ghave's first ability, Sage of Fables): the Shapers or another, and the
// card is drawn even if that creature then leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "744f79da-240f-4d14-a98c-e2796645ae88",
		Name:         "Shapers of Nature",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}{G}: Put a +1/+1 counter on target creature.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{3}{G}"),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.PlusOneCounterOnTheTarget,
			},
			{
				Label:   "{2}{U}, Remove a +1/+1 counter from a creature you control: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.RemoveCountersFrom(game.CounterPlusOne, 1, "a creature you control", effects.Creature())),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
		},
	})
}
