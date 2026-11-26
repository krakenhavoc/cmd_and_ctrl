package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undergrowth Scavenger — Creature — Fungus Horror {3}{G}, 0/0 (Magic 2015,
// #5556):
//
//	"This creature enters with a number of +1/+1 counters on it equal to the
//	 number of creature cards in all graveyards."
//
// Bone Devourer's entry counters (a CR 614.1c replacement counted as the
// creature enters): every creature card in every player's graveyard, the
// Scavenger itself not yet among them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42b17f78-d290-40ed-afac-f9362e526999",
		Name:         "Undergrowth Scavenger",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, _ *game.Card) int {
				return effects.B41CreatureCardsInAllGraveyards(g)
			}, "Undergrowth Scavenger enters with a number of +1/+1 counters equal to the number of creature cards in all graveyards"),
		},
	})
}
