package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frenetic Efreet — Creature — Efreet {1}{U}{R}, 2/1 (Mirage, S82 OLD b808):
//
//	"Flying
//	 {0}: Flip a coin. If you win the flip, this creature phases out. If you
//	 lose the flip, sacrifice this creature. (While it's phased out, it's
//	 treated as though it doesn't exist. It phases in before you untap during
//	 your next untap step.)"
//
// The flip is Crooked Scales' keyed call (CR 705): the controller calls it, and
// the result decides the branch. Winning phases the Efreet out through the
// phasing primitive (CR 702.26b); losing sacrifices it. Either branch does
// nothing if the Efreet has left or come back as a new object by then (CR
// 400.7), which happens when it was sacrificed in response.
//
// No simplification.
func init() {
	freneticEfreetFlipped = effects.CoinFlipThen("coin-flip/s82-old-b808-frenetic-efreet",
		func(ctx *effects.Context, r game.CoinFlipped) error {
			if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
				return nil
			}
			if effects.WonFirstFlip(r) {
				return effects.PhaseOut{Targets: []uuid.UUID{ctx.Item.SourceCardID}}.Apply(ctx)
			}
			return effects.SacrificePermanent{Target: ctx.Item.SourceCardID}.Apply(ctx)
		})
	effects.Register(effects.Spec{
		OracleID:        "c21d3360-8be0-477e-a592-7e64a76570d3",
		Name:            "Frenetic Efreet",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{0}: Flip a coin. If you win the flip, this creature phases out. If you lose the flip, sacrifice this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    effects.ManaCost("{0}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				g.FlipCoinForEffect(game.CoinFlipSpec{
					Flipper:  ctx.Controller(),
					Source:   ctx.Source(),
					Question: "Frenetic Efreet — call the coin flip",
					ThenKey:  freneticEfreetFlipped,
					Asked:    game.NewAskedItem(item),
				})
				return nil
			},
		}},
	})
}

var freneticEfreetFlipped game.CoinFlipThen
