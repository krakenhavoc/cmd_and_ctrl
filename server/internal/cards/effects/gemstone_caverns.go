package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gemstone Caverns — Legendary Land:
//
//	"If this card is in your opening hand and you're not the starting
//	 player, you may begin the game with Gemstone Caverns on the
//	 battlefield with a luck counter on it. If you do, exile a card
//	 from your hand.
//	 {T}: Add {C}. If Gemstone Caverns has a luck counter on it,
//	 instead add one mana of any color."
//
// Two sentences. The mana ability is one `ProducedFunc` reading the
// source's counters: a luck counter turns the colourless slot into a
// five-colour pick, and "instead" means it REPLACES the {C} rather than
// adding to it. The counter is read at ACTIVATION, not declared once,
// so a Caverns that gains a luck counter some other way (a Proliferate,
// a Vampire Hexmage removing one) answers correctly on the next tap.
//
// The opening-hand sentence is Spec.OpeningHand (CR 103.6a, ADR 0133).
// Every rider is its own option: NotTheStartingPlayer (the seat that
// takes the first turn is never asked), WithCounter("luck", 1) (the
// counter is on the land before anyone can act), ThenExileFromHand(1)
// (the owner chooses the card from what is left in their hand; a hand
// of nothing exiles nothing). A "no" leaves the Caverns in hand, an
// ordinary land with no luck counter that makes {C}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c0adbddc-b070-4c5f-afe0-0474c72a9251",
		Name:         "Gemstone Caverns",
		Completeness: CompletenessFull,
		OpeningHand: BeginTheGameOnTheBattlefield(
			NotTheStartingPlayer(),
			WithCounter("luck", 1),
			ThenExileFromHand(1),
		),
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: gemstoneCavernsProduced,
			Label:        "Add {C}, or one mana of any color with a luck counter",
		}},
	})
}

// gemstoneCavernsProduced is the "instead" clause: a five-colour pick
// while the land has a luck counter, plain {C} otherwise.
func gemstoneCavernsProduced(g *game.Game, _, source uuid.UUID) string {
	if c, ok := g.LookupCardForEffect(source); ok && c.Counters["luck"] > 0 {
		return "{W|U|B|R|G}"
	}
	return "{C}"
}
