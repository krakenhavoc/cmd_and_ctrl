package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sautekh Immortal — Artifact Creature — Necron {2}{B}, 2/2 (Warhammer 40,000
// Commander, #5948):
//
//	"Flash
//	 Elite Troops — This creature enters with a +1/+1 counter on it for each
//	 creature that died this turn."
//
// Grizzly Ghoul's CR 614.1c self-replacement over the table-wide tally of
// creatures that died this turn, anyone's, tokens included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7388a797-d40d-4783-a281-d17bc159b856",
		Name:            "Sautekh Immortal",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, _ *game.Card) int {
				return effects.B11CreaturesDiedThisTurn(g)
			}, "Sautekh Immortal enters with a +1/+1 counter on it for each creature that died this turn"),
		},
	})
}
