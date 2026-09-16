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
		OracleID:     "7e00b0cd-d212-4604-ba07-da21f4fe00b0",
		Name:         "Sram, Senior Edificer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && eventCardHasType(ev, g, "aura", "equipment", "vehicle")
			}, "Sram, Senior Edificer — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
