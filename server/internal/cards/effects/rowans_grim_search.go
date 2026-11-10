package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rowan's Grim Search — Instant {2}{B} (Wilds of Eldraine, #4471):
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 If this spell was bargained, look at the top four cards of your
//	 library, then put up to two of them back on top of your library in any
//	 order and the rest into your graveyard.
//	 You draw two cards and you lose 2 life."
//
// Bargain is the shared optional cost (g2Bargain). A bargained spell looks
// at the top four, asks which (up to two) stay (the pick is keyed, an
// on-disk identity, and carries the looked-at cards as its candidates),
// puts the rest into the graveyard (not milled), then orders the kept cards
// back on top (PutInLibraryInAnyOrder), and only then draws two and loses
// 2 life: the draw waits for the ordering, as the printed order says. An
// unbargained spell skips straight to the draw and the life loss. The life
// loss comes after the draws through the draw's own continuation, so a
// replacement prompt on the draw settles first.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "f60234e4-9b53-47a5-a21a-82c80b5af072",
		Name:          "Rowan's Grim Search",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Draws: 2},
		OptionalCosts: []game.AdditionalCost{g2Bargain()},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if !g2WasBargained(ctx) {
				return rowansGrimSearchDrawAndLose(ctx)
			}
			player := ctx.Controller()
			looked := ctx.Game.LookAtTopOfLibraryForEffect(player, 4)
			if len(looked) == 0 {
				return rowansGrimSearchDrawAndLose(ctx)
			}
			max := 2
			if len(looked) < max {
				max = len(looked)
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  player,
				Source:   ctx.Source(),
				Question: "Rowan's Grim Search — choose up to two of them to put back on top of your library (the rest go into your graveyard)",
				Cards:    looked,
				Min:      0,
				Max:      max,
				LiveZone: game.ZoneLibrary,
				ThenKey:  rowansGrimSearchPick,
				Asked:    game.NewAskedItem(ctx.Item),
			})
			return nil
		},
	})
}

// rowansGrimSearchPick puts the cards not kept into the graveyard, orders
// the kept ones on top, then draws and loses life. An on-disk key.
var rowansGrimSearchPick = askedPick("choose-cards/rowans-grim-search-pick", func(ctx *Context, r game.CardsChosen) error {
	kept := map[uuid.UUID]bool{}
	for _, id := range r.Picked {
		kept[id] = true
	}
	var rest []uuid.UUID
	for _, id := range r.Candidates {
		if !kept[id] {
			rest = append(rest, id)
		}
	}
	if err := restIntoGraveyard(ctx.Game, rest); err != nil {
		return err
	}
	return PutInLibraryInAnyOrder{
		Chooser:   ctx.Controller(),
		Cards:     r.Picked,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceTop,
		Label:     "Rowan's Grim Search — put them back on top of your library in any order",
		ThenTail:  ctx.ItemTail(rowansGrimSearchOrdered),
	}.Apply(ctx)
})

// rowansGrimSearchOrdered is the rest of the spell once the kept cards are
// back on top. An on-disk key: never renamed, never reused.
var rowansGrimSearchOrdered = CardTail("rowans-grim-search/ordered", func(ctx *Context, _ game.TailOutcome) error {
	return rowansGrimSearchDrawAndLose(ctx)
})

// rowansGrimSearchLoseLife is the draw's continuation: "and you lose 2
// life", after both cards are drawn. An on-disk key.
var rowansGrimSearchLoseLife = game.RegisterDrawThen("rowans-grim-search/lose-life", func(g *game.Game, d game.DrawThen) error {
	return g.ChangePlayerLifeForEffect(d.Source, d.Player, -2)
})

// rowansGrimSearchDrawAndLose is "You draw two cards and you lose 2 life."
func rowansGrimSearchDrawAndLose(ctx *Context) error {
	return ctx.Game.DrawNThenForEffect(ctx.Controller(), 2, game.DrawThen{
		Ref: rowansGrimSearchLoseLife, Player: ctx.Controller(), Source: ctx.Source(),
	})
}
