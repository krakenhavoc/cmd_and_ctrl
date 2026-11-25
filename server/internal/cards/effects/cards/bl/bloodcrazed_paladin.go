package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodcrazed Paladin — Creature — Vampire Knight {1}{B}, 1/1 (Ixalan,
// #5532):
//
//	"Flash
//	 This creature enters with a +1/+1 counter on it for each creature that
//	 died this turn."
//
// Bone Devourer's CR 614.1c self-replacement: the count is the table-wide
// tally of creatures that died this turn (anyone's), read as the Paladin
// enters. A Paladin cast with flash in response to a death counts it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7e7ec48c-e31d-4ddf-a128-1969951dacf9",
		Name:            "Bloodcrazed Paladin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, _ *game.Card) int {
				return effects.B11CreaturesDiedThisTurn(g)
			}, "Bloodcrazed Paladin enters with a +1/+1 counter on it for each creature that died this turn"),
		},
	})
}
