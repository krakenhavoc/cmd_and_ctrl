package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b159_helpers.go — shared bodies for Modern Horizons, batch 159 (#5519).

// OldB159DestroyThenItsControllerMaySearchForABasic is "Destroy target land.
// Its controller may search their library for a basic land card, put it onto
// the battlefield, then shuffle." (Geomancer's Gambit): Cleansing Wildfire's
// shape with the basic land entering untapped. The controller is read before
// the destroy, since the permanent is gone afterwards; a target that is no
// longer on the battlefield does nothing.
func OldB159DestroyThenItsControllerMaySearchForABasic(ctx *Context, target uuid.UUID, reason string) error {
	victim, ok := ControllerOfTarget(ctx, target)
	if !ok {
		return nil
	}
	if err := (DestroyTarget{Target: target}).Apply(ctx); err != nil {
		return err
	}
	return SearchLibrary{
		Player:    victim,
		Predicate: IsBasicLand,
		Dest:      game.ZoneBattlefield,
		Limit:     1,
		Shuffle:   true,
		Optional:  true,
		Reason:    reason,
	}.Apply(ctx)
}

// OldB159LookThreeMayShuffleThenDraw is "look at the top three cards of your
// library, then put them back in any order. You may shuffle. Draw a card."
// (Pondering Mage's enters trigger, Ponder's chain): the order is asked first,
// the shuffle second, and the draw follows on both answers (CR 608.2c).
func OldB159LookThreeMayShuffleThenDraw(ctx *Context) error {
	return LookAtTop{
		Player:   ctx.Controller(),
		N:        3,
		ThenTail: ctx.ItemTail(oldB159AskShuffle),
	}.Apply(ctx)
}

// oldB159AskShuffle is the look-at's "then": ask whether to shuffle. An
// on-disk key: never renamed, never reused.
var oldB159AskShuffle = CardTail("tail/old-b159-ask-shuffle", func(ctx *Context, _ game.TailOutcome) error {
	return MayChoice{
		Question: "Pondering Mage — shuffle your library?",
		YesLabel: "Shuffle",
		NoLabel:  "Keep that order",
		Then:     oldB159ShuffleThenDraw,
	}.Apply(ctx)
})

// oldB159ShuffleThenDraw is "you may shuffle", then the draw on both branches.
// An on-disk key: never renamed, never reused.
var oldB159ShuffleThenDraw = MayAnswer("confirm/old-b159-shuffle-then-draw", func(ctx *Context, r game.ConfirmAnswer) error {
	if r.Accepted {
		if err := ctx.Game.ShuffleLibraryForEffect(r.Chooser); err != nil {
			return err
		}
	}
	return ctx.Game.DrawNForEffect(r.Chooser, 1)
})
