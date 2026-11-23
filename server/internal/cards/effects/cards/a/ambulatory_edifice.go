package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ambulatory Edifice — Artifact Creature — Phyrexian Construct {2}{B}, 3/2:
//
//	"When this creature enters, you may pay 2 life. When you do, target
//	 creature gets -1/-1 until end of turn."
//
// The "you may pay" belongs to the enters trigger: it asks only a controller
// who can pay 2 life (CR 119.4, CR 119.8) and re-checks on the answer. A
// payment made raises a CR 603.12 reflexive trigger whose target is chosen
// as it goes on the stack, so it can be answered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b46351cf-73c7-493c-b8a7-5d7bedc8b205",
		Name:         "Ambulatory Edifice",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ambulatory Edifice — you may pay 2 life; when you do, target creature gets -1/-1 until end of turn", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if !effects.CanPayLifeNow(ctx, 2) {
					return nil
				}
				return effects.MayChoice{
					Question: "Ambulatory Edifice — pay 2 life?",
					YesLabel: "Pay 2 life",
					NoLabel:  "Don't pay",
					LifeCost: 2,
					Then:     effects.AmbulatoryEdificePaid,
				}.Apply(ctx)
			}),
		},
	})
}
