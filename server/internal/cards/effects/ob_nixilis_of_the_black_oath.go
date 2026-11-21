package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "cfc9a2f9-92fa-4608-b98d-f9763c89fc40",
		Name:            "Ob Nixilis of the Black Oath",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Emblem: &EmblemSpec{
			Label: "Ob Nixilis of the Black Oath emblem",
			Text:  "{1}{B}, Sacrifice a creature: You gain X life and draw X cards, where X is the sacrificed creature's power.",
			Activated: []ActivatedAbility{{
				Label:   "{1}{B}, Sacrifice a creature: You gain X life and draw X cards, where X is the sacrificed creature's power.",
				Cost:    Plus(ManaCost("{1}{B}"), SacrificeACreature()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect: Do(Instruction(func(ctx *Context) error {
					x := ctx.SacrificedPower()
					if x <= 0 {
						return nil
					}
					if err := (GainLife{Amount: x}).Apply(ctx); err != nil {
						return err
					}
					return DrawCards{N: x}.Apply(ctx)
				})),
			}},
		},
		Activated: []ActivatedAbility{
			{
				Label: "+2: Each opponent loses 1 life. You gain life equal to the life lost this way.",
				Cost:  LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return b21DrainEachOpponentAndGainTheTotal(g, item, 1)
				},
			},
			{
				Label:   "−2: Create a 5/5 black Demon creature token with flying. You lose 2 life.",
				Cost:    LoyaltyCost(-2),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect: Do(Instruction(func(ctx *Context) error {
					if err := (CreateToken{Template: TokenCard("5/5 black Demon with flying"), N: 1}).Apply(ctx); err != nil {
						return err
					}
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
				})),
			},
			{
				Label:  "−8: You get an emblem with \"{1}{B}, Sacrifice a creature: You gain X life and draw X cards, where X is the sacrificed creature's power.\"",
				Cost:   LoyaltyCost(-8),
				Effect: Do(CreateEmblem{}),
			},
		},
	})
}
