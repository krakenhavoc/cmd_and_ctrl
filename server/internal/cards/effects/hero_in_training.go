package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hero in Training — {2}{W} Creature — Human Hero 2/2 (Marvel Super Heroes,
// #4455):
//
//	"When this creature enters, draw a card. If you control another Hero,
//	 you gain 2 life."
//
// The Hero is looked for as the trigger resolves, after the draw: it is a
// branch of the effect, not an intervening "if", and it does not count this
// creature itself ("another"). A changeling counts as a Hero.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "87f1a840-021c-4699-a2c4-f2926daf98eb",
		Name:         "Hero in Training",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Hero in Training — draw a card; gain 2 life if you control another Hero", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
					return err
				}
				if !mshB4ControlsAnotherHero(g, item.Controller, item.SourceCardID, ctx.TextSubtype("Hero")) {
					return nil
				}
				return GainLife{Amount: 2}.Apply(ctx)
			}),
		},
	})
}
