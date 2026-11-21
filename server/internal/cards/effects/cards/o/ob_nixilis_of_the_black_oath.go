package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ob Nixilis of the Black Oath — Legendary Planeswalker — Nixilis
// {3}{B}{B}, loyalty 3 (Commander Masters, #5948):
//
//	"+2: Each opponent loses 1 life. You gain life equal to the life lost
//	 this way.
//	 −2: Create a 5/5 black Demon creature token with flying. You lose 2
//	 life.
//	 −8: You get an emblem with '{1}{B}, Sacrifice a creature: You gain X
//	 life and draw X cards, where X is the sacrificed creature's power.'
//	 Ob Nixilis of the Black Oath can be your commander."
//
// The +2 gains exactly what the opponents lost. The emblem's ability is an
// activated ability that functions from the command zone (CR 114.4): X is
// the sacrificed creature's power as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cfc9a2f9-92fa-4608-b98d-f9763c89fc40",
		Name:            "Ob Nixilis of the Black Oath",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Emblem: &effects.EmblemSpec{
			Label: "Ob Nixilis of the Black Oath emblem",
			Text:  "{1}{B}, Sacrifice a creature: You gain X life and draw X cards, where X is the sacrificed creature's power.",
			Activated: []effects.ActivatedAbility{{
				Label:   "{1}{B}, Sacrifice a creature: You gain X life and draw X cards, where X is the sacrificed creature's power.",
				Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.SacrificeACreature()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					x := ctx.SacrificedPower()
					if x <= 0 {
						return nil
					}
					if err := (effects.GainLife{Amount: x}).Apply(ctx); err != nil {
						return err
					}
					return effects.DrawCards{N: x}.Apply(ctx)
				})),
			}},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label: "+2: Each opponent loses 1 life. You gain life equal to the life lost this way.",
				Cost:  effects.LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.B21DrainEachOpponentAndGainTheTotal(g, item, 1)
				},
			},
			{
				Label:   "−2: Create a 5/5 black Demon creature token with flying. You lose 2 life.",
				Cost:    effects.LoyaltyCost(-2),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					if err := (effects.CreateToken{Template: effects.TokenCard("5/5 black Demon with flying"), N: 1}).Apply(ctx); err != nil {
						return err
					}
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
				})),
			},
			{
				Label:  "−8: You get an emblem with \"{1}{B}, Sacrifice a creature: You gain X life and draw X cards, where X is the sacrificed creature's power.\"",
				Cost:   effects.LoyaltyCost(-8),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}
