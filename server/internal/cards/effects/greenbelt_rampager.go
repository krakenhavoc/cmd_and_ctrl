package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c63a17b8-d183-48c7-bb04-edbfffba1e03",
		Name:         "Greenbelt Rampager",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Greenbelt Rampager — pay {E}{E}, or return it to hand and get {E}",
				func(g *game.Game, item *game.StackItem) error {
					return PayEnergyOrElse{
						N: 2,
						OrElse: func(ctx *Context) error {
							if sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
								if err := (BounceToHand{Target: ctx.Item.SourceCardID}).Apply(ctx); err != nil {
									return err
								}
							}
							return GetEnergy{N: 1}.Apply(ctx)
						},
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
