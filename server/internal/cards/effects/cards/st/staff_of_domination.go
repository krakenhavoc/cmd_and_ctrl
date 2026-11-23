package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Staff of Domination — Artifact {3} (EDHREC rank 1033):
//
//	"{1}: Untap this artifact.
//	 {2}, {T}: You gain 1 life.
//	 {3}, {T}: Untap target creature.
//	 {4}, {T}: Tap target creature.
//	 {5}, {T}: Draw a card."
//
// The Swiss-army combo piece: with a creature that taps for five or
// more it is infinite life, infinite cards, or an infinite untap.
// Five ordinary CR 602 activated abilities, each on the stack with a
// response window, in printed order. The two targeted ones use the
// ordinary "target creature" clause, so a hexproof creature can be
// neither tapped nor untapped, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7888719-647d-4022-a211-822fa09f0791",
		Name:         "Staff of Domination",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}: Untap this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{1}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{2}, {T}: You gain 1 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.GainLife{Player: item.Controller, Amount: 1}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{3}, {T}: Untap target creature.",
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					return effects.UntapTarget{Target: item.Targets[0].ID}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{4}, {T}: Tap target creature.",
				Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.TapChosenPermanent,
			},
			{
				Label:   "{5}, {T}: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
