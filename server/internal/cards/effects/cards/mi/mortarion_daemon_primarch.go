package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mortarion, Daemon Primarch — Legendary Creature — Demon Primarch
// {5}{B}, 5/6 (Warhammer 40,000 Commander, S82 GAP batch 188, #5948):
//
//	"Flying
//	 Primarch of the Death Guard — At the beginning of your end step, you
//	 may pay {X}. If you do, create X 2/2 black Astartes Warrior creature
//	 tokens with menace. X can't be greater than the amount of life you
//	 lost this turn."
//
// X is chosen as the ability resolves (CR 608.2d), from zero to the most you
// can pay, capped by the life the controller lost this turn (every loss,
// damage included); with none lost there is nothing to pay.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "588eecbf-dcfc-4b97-b81c-8b9f3d39e08f",
		Name:            "Mortarion, Daemon Primarch",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Mortarion, Daemon Primarch — you may pay {X} to create X 2/2 black Astartes Warrior tokens with menace (X up to the life you lost this turn)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					lost := effects.B18LifeLostThisTurn(g, ctx.Controller())
					if lost <= 0 {
						return nil
					}
					return effects.MayPayX{
						Cost:  "{X}",
						Max:   lost,
						Label: "Mortarion, Daemon Primarch",
						Buys:  "create X 2/2 black Astartes Warrior creature tokens with menace",
						Goal:  effects.XAsHighAsYouCan,
						Then:  mortarionPaid,
					}.Apply(ctx)
				}),
		},
	})
}

var mortarionPaid = effects.PayUnlessBranches("pay-unless/mortarion-warriors",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		return effects.CreateToken{Template: effects.TokenCard("2/2 black Astartes Warrior with menace"), N: r.Number}.Apply(ctx)
	}, nil)
