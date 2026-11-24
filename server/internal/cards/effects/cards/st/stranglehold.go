package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stranglehold — Enchantment {3}{R} (Secret Lair Drop, #5462):
//
//	"Your opponents can't search libraries.
//	 If an opponent would begin an extra turn, that player skips that turn
//	 instead."
//
// The first line is the engine's OpponentsCantSearchLibraries (CR 701.23a,
// game/search_cant.go): in the one search primitive, a search an opponent of
// Stranglehold's controller makes of a library finds nothing, the "then
// shuffle" the card prints still shuffles, and no search is recorded. The
// second is Trouble in Pairs' replacement of the extra turn at the moment it
// would begin (CR 614.10, SkipOpponentsExtraTurns): the controller's own extra
// turns are not touched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                     "b66c6e3a-daed-43a3-95d6-4f9d6223bbf1",
		Name:                         "Stranglehold",
		Completeness:                 effects.CompletenessFull,
		OpponentsCantSearchLibraries: true,
		Replacements:                 []game.ReplacementEffect{effects.SkipOpponentsExtraTurns()},
	})
}
