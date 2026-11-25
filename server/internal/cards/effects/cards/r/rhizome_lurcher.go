package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rhizome Lurcher — Creature — Fungus Zombie {2}{B}{G}, 2/2 (Guilds of
// Ravnica, #5525):
//
//	"Undergrowth — This creature enters with a number of +1/+1 counters on it
//	 equal to the number of creature cards in your graveyard."
//
// A replacement effect on its own entry (CR 614.1c): the creature cards are
// counted as it enters, after the spell has left the stack, so the Lurcher
// is not counted (it is on the battlefield, not in the graveyard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85e18895-d1d2-46c7-afd0-31da410c0438",
		Name:         "Rhizome Lurcher",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				return effects.B11CreatureCardsInGraveyard(g, src.Controller)
			}, "Rhizome Lurcher: enters with a +1/+1 counter for each creature card in your graveyard"),
		},
	})
}
