package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b06_helpers.go — shared pieces of S82 batch BNS-06 (#5929).

// bnsB06TriggeredAbilityItem passes for a TRIGGERED ability on the stack and
// fails for a spell or an activated ability: Consign to Memory's
// "triggered ability" half.
func bnsB06TriggeredAbilityItem() StackItemPredicate {
	return func(_ *game.Game, _ uuid.UUID, item *game.StackItem) bool {
		return item.Kind == game.StackItemTriggered
	}
}

// bnsB06PutRestOnBottomInAnyOrder is "put the rest on the bottom of your
// library in any order" as a rider of PutFromLibraryOntoBattlefield: the
// player orders the pile (the put_in_library prompt), unlike
// PutRestOnBottomInRandomOrderTail. An on-disk key: never renamed.
var bnsB06PutRestOnBottomInAnyOrder = PutFromLibraryTail("put-from-library/rest-on-bottom-in-any-order",
	func(ctx *Context, res PutFromLibraryResult) error {
		return ctx.Game.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
			Chooser:   res.Player,
			Source:    res.Source,
			Cards:     res.Rest,
			From:      game.ZoneLibrary,
			Placement: game.LibraryPlaceBottom,
			Reason:    libraryOrderLabel(ctx.Game, res.Source, "put the rest on the bottom of your library in any order"),
		})
	})
