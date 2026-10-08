package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Speed Demon — Legendary Creature — Demon {3}{B}{B}, 5/5:
//
//	"Flying, trample
//	 Start your engines!
//	 At the beginning of your end step, you draw X cards and lose X
//	 life, where X is your speed."
//
// ADR 0138 (#2122). X is read as the trigger resolves (CR 608.2h), off
// the controller's speed; a player with no speed draws nothing and
// loses nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c9fb22b4-dc22-4291-91f2-89c07cee7569",
		Name:            "The Speed Demon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample", StartYourEngines},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("The Speed Demon — draw X cards and lose X life, where X is your speed",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					x := YourSpeed(g, item.Controller)
					if x <= 0 {
						return nil
					}
					if err := (DrawCards{N: x}).Apply(ctx); err != nil {
						return err
					}
					return g.ChangePlayerLifeForEffect(ctx.Source(), item.Controller, -x)
				}),
		},
	})
}
