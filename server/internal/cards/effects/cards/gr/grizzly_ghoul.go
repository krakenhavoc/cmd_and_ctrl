package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grizzly Ghoul — Creature — Zombie Bear {2}{B}{G}, 4/3 (Midnight Hunt,
// #5497):
//
//	"Trample
//	 This creature enters with a +1/+1 counter on it for each creature that
//	 died this turn."
//
// Bone Devourer's CR 614.1c self-replacement, counted off the table-wide tally
// of creatures that died this turn (b11CreaturesDiedThisTurn, anyone's,
// tokens included).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "86783f8d-26c9-4b8c-a697-ab6facafeb67",
		Name:            "Grizzly Ghoul",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, _ *game.Card) int {
				return effects.B11CreaturesDiedThisTurn(g)
			}, "Grizzly Ghoul enters with a +1/+1 counter on it for each creature that died this turn"),
		},
	})
}
