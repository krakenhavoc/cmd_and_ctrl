package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b112_helpers.go — shapes Ikoria: Lair of Behemoths batch 112
// (#5512) prints: a few of the mutate triggers' bodies. "The number of times
// this creature has mutated" is batch 111's S82OldB111TimesMutated.

// S82OldB112LookAtTopLandToBattlefieldElseHand is Parcelbeast's ability: look
// at the top card of your library; if it's a land card you may put it onto the
// battlefield; if you don't put the card onto the battlefield, put it into
// your hand (so a nonland card, or a land you decline, goes to your hand).
func S82OldB112LookAtTopLandToBattlefieldElseHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	top := g.LookAtTopOfLibraryForEffect(item.Controller, 1)
	if len(top) == 0 {
		return nil
	}
	if c, ok := g.LookupCardForEffect(top[0]); ok && c.IsLand() {
		return PutFromLibraryOntoBattlefield{
			Player:   item.Controller,
			Cards:    top,
			Match:    Land(),
			Max:      1,
			Optional: true,
			Label:    "Parcelbeast — you may put the land onto the battlefield; otherwise it goes into your hand",
			ThenTail: S82OldB112LandNotPutToHand,
		}.Apply(ctx)
	}
	return g.TakeFromLibraryToHandKeyedForEffect(item.Controller, top, game.KeyedTail{})
}

// S82OldB112LandNotPutToHand is the rest of Parcelbeast's look: a land that
// was not put onto the battlefield goes into your hand. An on-disk key: never
// renamed, never reused.
var S82OldB112LandNotPutToHand = PutFromLibraryTail("put-from-library/s82-old-b112-parcelbeast-hand",
	func(ctx *Context, res PutFromLibraryResult) error {
		if len(res.Entered) > 0 || len(res.Rest) == 0 {
			return nil
		}
		return ctx.Game.TakeFromLibraryToHandKeyedForEffect(res.Player, res.Rest[:1], game.KeyedTail{})
	})
