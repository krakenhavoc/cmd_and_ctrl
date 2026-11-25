package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_color_of_choice.go — #6210: "if <a land> is tapped for mana, it
// produces mana of a color of your choice instead of any other type"
// (Pulse of Llanowar) and "… one mana of a color of your choice instead
// of any other type and amount" (Harvest Mage). The colour is chosen as
// part of the tap (game/mana_color_of_choice.go): a tap by hand asks the
// five-colour pick Birds of Paradise asks, and the auto-tapper picks the
// colour the cost needs.
//
// Append-only: add a builder, never change what one means.

// TappedForManaProducesAColorOfYourChoice is the replacement. `applies`
// says which tapped land it covers (it is asked only for a production
// that is a tap for mana, CR 106.12a, and makes at least one mana);
// `oneMana` is Harvest Mage's "one mana … instead of any other type and
// amount", and false keeps the amount (Pulse of Llanowar). "Your choice"
// is the player tapping the land, whose pool the mana reaches.
func TappedForManaProducesAColorOfYourChoice(label string, oneMana bool, applies func(land game.Card, ev *game.ReplacementEvent, g *game.Game, source *game.Card) bool) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventManaAdded},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, source *game.Card) bool {
			if ev.Kind != game.RepEventProduceMana || !ev.ManaFromTap || len(ev.ManaColors) == 0 {
				return false
			}
			land, ok := g.LookupCardForEffect(ev.ManaSource)
			return ok && land.IsLand() && applies(land, ev, g, source)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			n := len(ev.ManaColors)
			if oneMana {
				n = 1
			}
			out := make([]string, n)
			for i := range out {
				out[i] = game.ManaOfChosenColor
			}
			ev.ManaColors = out
			return nil
		},
		Controller: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) uuid.UUID {
			return ev.ManaPlayer
		},
		Label: label,
	}
}

// BasicLandYouControl is Pulse of Llanowar's "a basic land you control":
// the land is basic (CR 205.4c) and controlled by the replacement's
// source's controller.
func BasicLandYouControl(land game.Card, _ *game.ReplacementEvent, _ *game.Game, source *game.Card) bool {
	return source != nil && land.Controller == source.Controller && land.HasSupertype("basic")
}
