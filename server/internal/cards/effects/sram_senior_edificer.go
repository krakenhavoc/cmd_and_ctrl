package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sram, Senior Edificer — Legendary Creature — Dwarf Advisor {1}{W},
// 2/2 (EDHREC rank 442):
//
//	"Whenever you cast an Aura, Equipment, or Vehicle spell, draw a
//	 card."
//
// The Voltron deck's card draw, and a card whose payoff is ahead of
// its subject: attachments themselves are #280 (S33), but CASTING
// an Aura or Equipment is an ordinary cast event today, and the
// trigger reads the spell's printed subtypes off the stack. When the
// attachment machinery lands, Sram needs no change.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "7e00b0cd-d212-4604-ba07-da21f4fe00b0",
		Name:     "Sram, Senior Edificer",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && b03CastSpellHasType(ev, g, "aura", "equipment", "vehicle")
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Sram, Senior Edificer — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
