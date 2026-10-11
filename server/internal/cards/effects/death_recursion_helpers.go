package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// death_recursion_helpers.go — the shared bodies of the death, discard
// and recursion cards of the All Will Be Black request (#2952).
// Append-only.

// returnSelfFromGraveyardToHand is "return it to its owner's hand" for a
// card whose own "when this dies / is put into a graveyard from the
// battlefield" trigger has the source in a graveyard now: Endless
// Cockroaches and Glistening Oil.
//
// "It" is the object that died (CR 400.7). If the card has left the
// graveyard since, and perhaps come back, it is a new object the
// ability cannot find, so nothing happens: sameGraveyardObject compares
// the object epoch the trigger stamped. A token ceased to exist in the
// graveyard (CR 111.7) and is not found either.
func returnSelfFromGraveyardToHand(g *game.Game, item *game.StackItem) error {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard || !sameGraveyardObject(z, item) {
		return nil
	}
	return ReturnFromGraveyard{Target: item.SourceCardID, Dest: game.ZoneHand}.Apply(NewContext(g, item))
}
