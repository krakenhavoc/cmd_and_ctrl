package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fifth_dawn_helpers.go — shared bodies for the Fifth Dawn cards (#5604).

// shuffleThisSpellIntoItsOwnersLibrary is the Beacons' last sentence,
// "Shuffle <this> into its owner's library": the resolving card is put
// into its owner's library, which is then shuffled, instead of going to
// the graveyard (Beacon of Unrest's tail). A Beacon that is countered,
// or that has no legal target left, never gets here and goes to the
// graveyard as usual (CR 608.2b).
func shuffleThisSpellIntoItsOwnersLibrary(ctx *Context) error {
	return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{ctx.Source()}, game.TuckOptions{},
		game.KeyedTail{Ref: thenShuffle, Player: ctx.Item.Owner})
}
