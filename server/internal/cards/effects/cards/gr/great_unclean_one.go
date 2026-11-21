package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Great Unclean One — Creature — Demon {4}{B}, 4/5 (Warhammer 40,000
// Commander, S82 GAP batch 187, #5948):
//
//	"Reverberating Summons — At the beginning of your end step, each
//	 opponent loses 2 life. Then for each opponent who has less life than
//	 you, create a 1/3 black Demon creature token named Plaguebearer of
//	 Nurgle."
//
// "Then" reads the life totals once the loss is done, as the ability
// resolves (CR 608.2c): one token for each opponent still in the game whose
// life is below the controller's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd9031e4-9f6f-4dd8-908e-dcf7b79438e5",
		Name:         "Great Unclean One",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Great Unclean One — each opponent loses 2 life, then for each opponent with less life than you, create a Plaguebearer of Nurgle",
				func(g *game.Game, item *game.StackItem) error {
					if err := effects.EachOpponentLosesLifeNow(g, item, 2); err != nil {
						return err
					}
					ctx := effects.NewContext(g, item)
					me := g.PlayerByIDForEffect(ctx.Controller())
					if me == nil {
						return nil
					}
					n := 0
					for _, opp := range ctx.Opponents() {
						if p := g.PlayerByIDForEffect(opp); p != nil && !p.Eliminated && p.Life < me.Life {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("1/3 black Demon named Plaguebearer of Nurgle"), N: n}.Apply(ctx)
				}),
		},
	})
}
