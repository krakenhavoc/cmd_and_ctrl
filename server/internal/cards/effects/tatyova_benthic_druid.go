package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tatyova, Benthic Druid — Legendary Creature — Merfolk Druid
// {3}{G}{U}, 3/3 (EDHREC rank 472):
//
//	"Landfall — Whenever a land you control enters, you gain 1 life
//	 and draw a card."
//
// The Simic lands commander: every land drop, fetch and ramp spell
// is a card. Landfall is the ETB trigger filtered to lands, firing
// for any way a land enters under your control (Tireless
// Provisioner's shape); a fetchland fires it twice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0715e860-3b3b-4331-9718-207973e94fee",
		Name:         "Tatyova, Benthic Druid",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsLand()
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Tatyova — gain 1 life and draw a card (landfall)",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if err := (GainLife{Player: item.Controller, Amount: 1}).Apply(ctx); err != nil {
							return err
						}
						return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
					})
			},
		}},
	})
}
