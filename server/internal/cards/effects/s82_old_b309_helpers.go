package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b309_helpers.go — Battle for Zendikar landfall cards (#5547).

// S82OldB309EnteredLandIs reports whether the land that entered (the
// landfall trigger's event card) has the subtype, read through the text
// reader so a changed land type is honoured, and from last-known
// information when the land has since left.
func S82OldB309EnteredLandIs(g *game.Game, item *game.StackItem, subtype string) bool {
	return RfCreatureCEnteredLandHasSubtype(g, item, g.TextReaderForEffect(item).TextSubtype(subtype))
}
