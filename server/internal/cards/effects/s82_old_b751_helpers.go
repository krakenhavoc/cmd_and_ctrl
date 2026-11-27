package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b751_helpers.go — shared vocabulary for the S82 OLD batch 751 cards
// (Stronghold, #5627).

// OldB751SearchForTheNameOf is "Search your library for a card with the same
// name as <that creature>, put that card onto the battlefield, then shuffle"
// (Mask of the Mimic): one card of exactly that name goes onto the battlefield
// under the searcher's control. Only a permanent card can go there, so the
// search offers just those.
func OldB751SearchForTheNameOf(ctx *Context, card, name string) error {
	return SearchLibrary{
		Player:    ctx.Controller(),
		Predicate: func(l game.Card) bool { return l.Name == name && l.IsPermanent() },
		Dest:      game.ZoneBattlefield,
		Limit:     1,
		Shuffle:   true,
		Reason:    card + " — a card named " + name,
	}.Apply(ctx)
}

// OldB751ReturnTheObjectsToHandsBody is the delayed trigger "return <the objects
// named in Params.Objects> to their owners' hands" (Contempt: the enchanted
// creature and the Aura): each one returns only if it is still the object it was
// when the trigger was created and is on the battlefield (CR 400.7). Its key is
// on disk.
var OldB751ReturnTheObjectsToHandsBody = game.SimpleDelayedBody("s82-old-b751/return-the-objects-to-hands", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, obj := range item.Params.Objects {
		if !OnBattlefield(g, obj.ID) || g.ObjectEpochForEffect(obj.ID) != obj.Epoch {
			continue
		}
		if err := (BounceToHand{Target: obj.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// OldB751TwoOrMoreOtherCreatures reports whether, besides `except`, at least
// two creatures are on the battlefield (Portcullis's intervening "if").
func OldB751TwoOrMoreOtherCreatures(g *game.Game, except uuid.UUID) bool {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != except && c.IsCreature() {
			n++
		}
	}
	return n >= 2
}
