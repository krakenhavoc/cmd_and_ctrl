package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Greenbelt Rampager — Creature — Elephant {G}, 3/4:
//
//	"When this creature enters, pay {E}{E} (two energy counters). If you
//	 can't, return this creature to its owner's hand and you get {E}."
//
// ADR 0129 §3 (#1995): the payment is mandatory (CR 118.12, "started to
// pay a mandatory cost"), so nothing is asked. A controller with two
// energy pays them; one without pays nothing (CR 118.3), the Elephant
// goes back to its owner's hand if it is still the permanent that
// entered (CR 400.7), and they get {E}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c63a17b8-d183-48c7-bb04-edbfffba1e03",
		Name:         "Greenbelt Rampager",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Greenbelt Rampager — pay {E}{E}, or return it to hand and get {E}",
				func(g *game.Game, item *game.StackItem) error {
					return effects.PayEnergyOrElse{
						N: 2,
						OrElse: func(ctx *effects.Context) error {
							if effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
								if err := (effects.BounceToHand{Target: ctx.Item.SourceCardID}).Apply(ctx); err != nil {
									return err
								}
							}
							return effects.GetEnergy{N: 1}.Apply(ctx)
						},
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
