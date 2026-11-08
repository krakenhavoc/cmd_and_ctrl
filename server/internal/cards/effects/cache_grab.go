package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cache Grab — Instant {1}{G} (Bloomburrow, #4467):
//
//	"Mill four cards. You may put a permanent card from among the cards
//	 milled this way into your hand. If you control a Squirrel or returned
//	 a Squirrel card to your hand this way, create a Food token."
//
// Something Worth Saving's mill and pick (rfSpellBMillThenMayTakePermanent):
// the pick is over the cards that actually reached the graveyard (CR
// 400.7), up to one, so declining is an answer. Once it is answered the
// Food is made if a Squirrel card was put into your hand this way (it is
// in your hand now), or if you control a Squirrel as the clause is read.
// "Squirrel" is a word of the card's text (CR 612.2), read through it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "06ccec61-27d2-4079-80b5-29ebd3a980ee",
		Name:         "Cache Grab",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return rfSpellBMillThenMayTakePermanent(ctx,
				"Cache Grab — put a permanent card milled this way into your hand",
				cacheGrabMilled)
		},
	})
}

// cacheGrabMilled is the mill's keyed "then": the pick, then the Food
// clause told what was taken. An on-disk key: never renamed, never reused.
var cacheGrabMilled = rfSpellBTakeFromMilledTailPicked("mill-then-take/cache-grab", func(ctx *Context, picked []uuid.UUID) error {
	squirrel := ctx.TextSubtype("Squirrel")
	if !cacheGrabReturnedSquirrel(ctx, picked, squirrel) && !cacheGrabControlsSquirrel(ctx, squirrel) {
		return nil
	}
	return CreateToken{Template: FoodToken(), N: 1}.Apply(ctx)
})

// cacheGrabReturnedSquirrel is "returned a Squirrel card to your hand this
// way": the picked card reached your hand and is a Squirrel card.
func cacheGrabReturnedSquirrel(ctx *Context, picked []uuid.UUID, squirrel string) bool {
	for _, id := range picked {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneHand {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.HasSubtype(squirrel) {
			return true
		}
	}
	return false
}

// cacheGrabControlsSquirrel is "you control a Squirrel".
func cacheGrabControlsSquirrel(ctx *Context, squirrel string) bool {
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.HasSubtype(squirrel) {
			return true
		}
	}
	return false
}
