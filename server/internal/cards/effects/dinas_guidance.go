package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dina's Guidance — Instant {1}{B}{G} (Secrets of Strixhaven, #4456):
//
//	"Search your library for a creature card, reveal it, put it into your
//	 hand or graveyard, then shuffle."
//
// Dark Tutor's shape. The search finds the card and reveals it to the table
// (CR 701.23); it stays in the library, on top, while its finder chooses
// hand or graveyard with it in view, and the shuffle comes after the move
// as printed. A search that finds nothing (no creature card, or the finder
// declines to find one: a stated quality allows it, CR 701.23b) still
// shuffles. A card that is no longer where the search left it when the
// choice is made is skipped, and the shuffle still happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a9abed25-9256-4fc7-8afb-edb4286e18f4",
		Name:         "Dina's Guidance",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(c game.Card) bool { return c.IsCreature() },
				Dest:      game.ZoneLibrary,
				ToTop:     true,
				Limit:     1,
				Reveal:    true,
				Shuffle:   false,
				Reason:    "Dina's Guidance — search your library for a creature card",
				ThenTail:  ctx.Tail(dinasGuidanceFound),
			}.Apply(ctx)
		},
	})
}

// dinasGuidanceFound is the search's continuation. The found card is still
// in the library, on top, so the choice of where it goes (hand or
// graveyard) is made with it in view. Nothing found: the shuffle still
// happens. An on-disk key.
var dinasGuidanceFound = CardTail("dinas-guidance/found", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
	}
	return PickOption{
		Question: "Dina's Guidance — put the card into your hand or your graveyard?",
		Options:  []game.ChoiceOption{{Label: "Hand"}, {Label: "Graveyard"}},
		ThenKey:  dinasGuidancePlaced,
		Carry:    o.IDs,
	}.Apply(ctx)
})

// dinasGuidancePlaced puts the found card where the player chose, then
// shuffles. A card that is no longer where the search left it is skipped.
// An on-disk key.
var dinasGuidancePlaced = OptionPickThen("option-pick/dinas-guidance-placed", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || ctx.Game.FindCardZoneForEffect(r.Carry[0]) == nil {
		return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
	}
	if r.Index == 1 {
		if err := ctx.Game.PutIntoGraveyardForEffect(r.Carry[0]); err != nil {
			return err
		}
		return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
	}
	return ctx.Game.BounceCardsToHandKeyedForEffect(liveCards(ctx.Game, r.Carry[:1]), ctx.Tail(dinasGuidanceAfterHand))
})

// dinasGuidanceAfterHand is the tail of the move to hand: the shuffle.
var dinasGuidanceAfterHand = CardTail("dinas-guidance/after-hand", func(ctx *Context, _ game.TailOutcome) error {
	return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
})
