package protocol

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/oracletext"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// abilityTextOf is StackItemView.AbilityText (#3830): the printed text
// of the ability an item is, or "" when the table shows its label.
//
// The text is the catalog row's own clause on its card, so only an item
// that names a row of its source's own definition has one: the two
// catalog bodies stamp that ref (ADR 0041 P9). A spell, a copy (the
// owner's call on #3830, though a copy has its original's text), a
// granted row, a delayed, reflexive or rule-made trigger, and a
// legacy closure have none.
//
// And it is never a way to read a hidden card. A source that is face
// down, or sits in a hand or library some seat cannot see, has no text
// here, for every viewer, its controller included. So does a source
// whose text is no longer the row's: a borrowed row (the lender's
// text), or a permanent that exchanged its text box or stopped being
// the copy whose row it was. A source that has left the game entirely
// (a token) was public when it did, and keeps its text.
//
// Caller holds the game's lock, as every view read does.
func abilityTextOf(g *game.Game, it *game.StackItem) string {
	if it.Kind == game.StackItemSpell || it.IsCopy || it.GrantedBy != uuid.Nil || it.Params.Ability == nil {
		return ""
	}
	if it.Body != game.CatalogActivatedBodyKey && it.Body != game.CatalogTriggeredBodyKey {
		return ""
	}
	ref := *it.Params.Ability
	if c, ok := g.LookupCardPtrForEffect(it.SourceCardID); ok {
		if c.FaceDown || game.BaseCatalogKey(game.CatalogKeyOf(c)) != game.BaseCatalogKey(ref.Key) {
			return ""
		}
		if z := g.FindCardZoneForEffect(it.SourceCardID); z != nil &&
			(z.Kind == game.ZoneHand || z.Kind == game.ZoneLibrary) && !knownToEveryPlayer(g, c) {
			return ""
		}
	}
	return oracletext.AbilityText(ref)
}

// knownToEveryPlayer: every seat knows the card (a revealed ninjutsu
// or miracle card in a hand).
func knownToEveryPlayer(g *game.Game, c *game.Card) bool {
	if len(g.Seats) == 0 {
		return false
	}
	for _, p := range g.Seats {
		if p != nil && !c.KnownBy[p.ID] {
			return false
		}
	}
	return true
}
