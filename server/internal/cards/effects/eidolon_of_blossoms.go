package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eidolon of Blossoms — Enchantment Creature — Spirit {2}{G}{G}, 2/2
// (EDHREC rank 894):
//
//	"Constellation — Whenever this creature or another enchantment
//	 you control enters, draw a card."
//
// The enchantress deck's other engine, and it counts itself: the
// Eidolon is an enchantment, so its own entry draws. Sun Titan's
// "one ability, two conditions" shape on a single event kind — the
// source's own ETB, or another enchantment entering under the same
// control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77ccbea1-70af-4194-adad-39a904221c75",
		Name:         "Eidolon of Blossoms",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return true
				}
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.IsEnchantment()
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Eidolon of Blossoms — draw a card (constellation)",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
