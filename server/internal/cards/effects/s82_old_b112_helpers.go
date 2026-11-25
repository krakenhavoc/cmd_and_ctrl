package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b112_helpers.go — shapes Ikoria: Lair of Behemoths batch 112
// (#5512) prints: "the number of times this creature has mutated" (as batch
// 111 reads it, under this batch's own name) and a few of the mutate
// triggers' bodies.

// S82OldB112TimesMutated is "the number of times this creature has mutated"
// (Archipelagore, Porcuparrot): each merge adds one component to the
// permanent, so a permanent made of N cards has mutated N-1 times (CR
// 702.140d, CR 730.2). A permanent that has left the battlefield is counted
// from the turn's log of merges into it, since its last entry; the answer is
// at least one, since the merge that triggered the ability counts.
func S82OldB112TimesMutated(g *game.Game, id uuid.UUID) int {
	if c, ok := g.LookupCardForEffect(id); ok && OnBattlefield(g, id) && c.IsMerged() {
		return max(len(c.MergedFrom)-1, 1)
	}
	n := 0
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.CardID != id {
			continue
		}
		if ev.Kind == game.EventMutate {
			n++
		}
		if ev.Kind == game.EventZoneMove && ev.NewZone == game.ZoneBattlefield {
			break
		}
	}
	return max(n, 1)
}

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
