package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b98_helpers.go — shared bodies for GAP batch 98 (#5948).

// gapB98AbilityFromColorlessSource is "…ability you control from a colorless
// source" (Abstruse Archaic): the ability's source as it is now, or as it
// last was on the battlefield when it has left (a sacrificed artifact's
// ability is still from that colorless artifact, CR 608.2h), has no color.
// A source that is a card in another zone (a cycling ability) is read as the
// card. A source that can be found nowhere fails the clause.
func gapB98AbilityFromColorlessSource() AbilityPredicate {
	return func(g *game.Game, chooser uuid.UUID, item *game.StackItem) bool {
		if item.Controller != chooser {
			return false
		}
		if ref, ok := g.SourceObjectForEffect(item); ok {
			if info, ok := g.PermanentForEffect(ref); ok {
				return len(info.Characteristic.Colors) == 0
			}
		}
		c, ok := g.LookupCardForEffect(item.SourceCardID)
		return ok && c.IsColorless()
	}
}

// gapB98ExileChosen is the registered answer to Bane of Bala Ged's pick: the
// defending player exiles each permanent they chose. An on-disk key: never
// renamed, never reused.
var gapB98ExileChosen = game.RegisterChooseCardsThen("choose-cards/gap-b98-exile-chosen-permanents", exileEachChosenCard)
