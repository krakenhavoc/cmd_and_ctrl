package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Make Your Own Luck — Sorcery {3}{G}{U} (Outlaws of Thunder Junction,
// #4468):
//
//	"Look at the top three cards of your library. You may exile a
//	 nonland card from among them. If you do, it becomes plotted. Put the
//	 rest into your hand. (You may cast it as a sorcery on a later turn
//	 without paying its mana cost.)"
//
// One choose_cards prompt over the nonland cards among the three, up to
// one. The chosen card is exiled with a keyed tail and plotted by Aven
// Interrupter's PlotExiled, only if it reached exile ("if you do"), and
// the plot's own grant lets you cast it free on a later turn
// (CR 702.170d). The rest of the three, lands included, then go to your
// hand: all three when you exile nothing or none of them is a nonland
// card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "41246140-4b44-40c4-b92b-501e49c50bf7",
		Name:         "Make Your Own Luck",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			looked := ctx.Game.LookAtTopOfLibraryForEffect(player, 3)
			var nonland []uuid.UUID
			for _, id := range looked {
				if c, ok := ctx.Game.LookupCardForEffect(id); ok && !c.IsLand() {
					nonland = append(nonland, id)
				}
			}
			if len(nonland) == 0 {
				ctx.Game.TakeFromLibraryToHandForEffect(player, looked)
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  player,
				Source:   ctx.Source(),
				Question: "Make Your Own Luck — you may exile a nonland card from among them; it becomes plotted (the rest go to your hand)",
				Cards:    nonland,
				Min:      0,
				Max:      1,
				LiveZone: game.ZoneLibrary,
				ThenKey:  makeYourOwnLuckPick,
				Carry:    looked,
				Asked:    game.NewAskedItem(ctx.Item),
			})
			return nil
		},
	})
}

// makeYourOwnLuckPick is the answer: exile and plot the chosen card, or
// put all three into your hand.
var makeYourOwnLuckPick = askedPick("choose-cards/make-your-own-luck-plot", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		ctx.Game.TakeFromLibraryToHandForEffect(ctx.Controller(), r.Carry)
		return nil
	}
	picked := r.Picked[0]
	var rest []uuid.UUID
	for _, id := range r.Carry {
		if id != picked {
			rest = append(rest, id)
		}
	}
	tail := ctx.ItemTail(makeYourOwnLuckPlotThenRest)
	tail.Carry = rest
	return ExileTarget{Target: picked, ThenTail: tail}.Apply(ctx)
})

// makeYourOwnLuckPlotThenRest is "if you do, it becomes plotted. Put the
// rest into your hand."
var makeYourOwnLuckPlotThenRest = CardTail("make-your-own-luck/plot-then-rest", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := (PlotExiled{Card: id}).Apply(ctx); err != nil {
			return err
		}
	}
	ctx.Game.TakeFromLibraryToHandForEffect(ctx.Controller(), o.Tail.Carry)
	return nil
})
