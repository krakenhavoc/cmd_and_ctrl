package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stillness in Motion — Enchantment {1}{U}:
//
//	"At the beginning of your upkeep, mill three cards. Then if your
//	 library has no cards in it, exile this enchantment and put five
//	 cards from your graveyard on top of your library in any order."
//
// The mill runs first and the library is checked once it has landed
// (MillToZone.Then, so a commander's CR 903.9 question is answered
// before the count). The check is not an intervening "if" (CR 603.4):
// the trigger always goes on the stack, and the condition is read on
// resolution after the mill.
//
// "This enchantment" is exiled only while it is still the permanent the
// ability came from (CR 400.7); the five cards are put back either way.
// The controller chooses which five and their order (ADR 0088's
// put_in_library prompt, the first chosen on top). With five or fewer
// cards in the graveyard every one goes back and only the order is
// asked (CR 608.2 — do as much as possible).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d73913ea-44d0-408f-9f8b-8e91843f2826",
		Name:         "Stillness in Motion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Stillness in Motion — mill three cards, then if your library is empty, exile this and put five cards from your graveyard on top", stillnessInMotionUpkeep),
		},
	})
}

const stillnessInMotionReturned = 5

func stillnessInMotionUpkeep(g *game.Game, item *game.StackItem) error {
	return MillToZone{N: 3, Then: func(ctx *Context, _ []uuid.UUID) error {
		p := ctx.PlayerByID(ctx.Controller())
		if p == nil || (p.Library != nil && p.Library.Size() > 0) {
			return nil
		}
		return ExileTarget{Target: ctx.Item.SourceCardID, Then: func(ctx *Context, _ bool) error {
			return stillnessInMotionRestack(ctx)
		}}.Apply(ctx)
	}}.Apply(NewContext(g, item))
}

// stillnessInMotionRestack is "put five cards from your graveyard on top
// of your library in any order".
func stillnessInMotionRestack(ctx *Context) error {
	player, source := ctx.Controller(), ctx.Source()
	yard := graveyardCardIDs(ctx, player, func(game.Card) bool { return true })
	if len(yard) == 0 {
		return nil
	}
	label := "Stillness in Motion — put five cards from your graveyard on top of your library"
	order := func(g *game.Game, picked []uuid.UUID) error {
		return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
			Chooser:   player,
			Source:    source,
			Cards:     picked,
			From:      game.ZoneGraveyard,
			Placement: game.LibraryPlaceTop,
			Reason:    label + " — in what order? (the first is on top)",
		})
	}
	if len(yard) <= stillnessInMotionReturned {
		return order(ctx.Game, yard)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   source,
		Question: label,
		Cards:    yard,
		Min:      stillnessInMotionReturned,
		Max:      stillnessInMotionReturned,
		Zone:     game.ZoneGraveyard,
		Then:     order,
	})
	return nil
}
