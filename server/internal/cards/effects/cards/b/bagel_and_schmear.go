package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bagel and Schmear — Artifact — Food {1} (Marvel's Spider-Man):
//
//	"Share — {W}, {T}, Sacrifice this artifact: Put a +1/+1 counter on up
//	 to one target creature. Draw a card. Activate only as a sorcery.
//	 Nosh — {2}, {T}, Sacrifice this artifact: You gain 3 life and draw a
//	 card."
//
// Two Food abilities (CR 111.10b names Food's sacrifice; the card prints
// both). The target of Share is "up to one": chosen or not, the draw
// happens, and a target that has gone by resolution only loses the counter
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1389c9fa-fec2-4212-a7aa-6dfb318dc202",
		Name:         "Bagel and Schmear",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:        "Share — {W}, {T}, Sacrifice this artifact: Put a +1/+1 counter on up to one target creature. Draw a card. Activate only as a sorcery.",
				Cost:         effects.Plus(effects.ManaCost("{W}"), effects.TapCost(), effects.SacrificeThis()),
				SorcerySpeed: true,
				Targets:      effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if err := effects.PlusOneCounterOnTheTarget(g, item); err != nil {
						return err
					}
					return effects.DrawCards{N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "Nosh — {2}, {T}, Sacrifice this artifact: You gain 3 life and draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.GainLife{Player: item.Controller, Amount: 3}).Apply(ctx); err != nil {
						return err
					}
					return effects.DrawCards{N: 1}.Apply(ctx)
				},
			},
		},
	})
}
