package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vedalken Plotter — Creature — Vedalken Wizard {2}{U}, 1/1 (Commander 2011,
// #5574):
//
//	"When this creature enters, exchange control of target land you control
//	 and target land an opponent controls."
//
// Two target clauses, each judged on its own at resolution. The exchange is
// all or nothing (CR 701.12a): if either land is no longer a legal target, or
// has changed controller so the pair no longer fits, nothing moves. It has no
// duration.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f85b92d-7212-4fc7-a5e8-6350e1acfdd3",
		Name:         "Vedalken Plotter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Vedalken Plotter — exchange control of target land you control and target land an opponent controls",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					mine, okMine := ctx.ClauseTarget(0)
					theirs, okTheirs := ctx.ClauseTarget(1)
					if !okMine || !okTheirs || mine.Kind != game.TargetCard || theirs.Kind != game.TargetCard {
						return nil
					}
					return effects.ExchangeControl{A: mine.ID, B: theirs.ID, Label: "Vedalken Plotter — exchange control of two lands"}.Apply(ctx)
				}),
				effects.Clauses(
					effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl()),
					effects.TargetPermanent("target land an opponent controls", effects.Land(), effects.OpponentControls()),
				)),
		},
	})
}
