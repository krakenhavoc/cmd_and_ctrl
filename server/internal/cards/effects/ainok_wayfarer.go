package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ainok Wayfarer — Creature — Dog Scout {1}{G}, 1/1 (Tarkir: Dragonstorm,
// #4463):
//
//	"When this creature enters, mill three cards. You may put a land card
//	 from among them into your hand. If you don't, put a +1/+1 counter on
//	 this creature."
//
// Six's mill-then-take shape (mill_then_take.go), filtered to lands, with
// the "if you don't" branch written into the pick's continuation: no land
// taken, whether because none was milled or because the player declined,
// is the counter. Only the cards this mill put into the graveyard are
// offered. A Wayfarer that left the battlefield in response gets no
// counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8134f3fd-197a-4b44-a370-311cc955dea3",
		Name:         "Ainok Wayfarer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ainok Wayfarer — mill three cards; you may put a land card from among them into your hand, or put a +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{N: 3, ThenTail: ctx.ItemTail(ainokWayfarerMilled)}.Apply(ctx)
				}),
		},
	})
}

// ainokWayfarerMilled offers a land from among the milled cards, or puts
// the counter on when there is none to offer. An on-disk key.
var ainokWayfarerMilled = CardTail("ainok-wayfarer/after-mill", func(ctx *Context, o game.TailOutcome) error {
	player := ctx.Controller()
	var lands []uuid.UUID
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !Land()(ctx.Game, player, c) {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		lands = append(lands, id)
	}
	if len(lands) == 0 {
		return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Ainok Wayfarer — you may put a land card from among them into your hand (if you don't, put a +1/+1 counter on it)",
		Cards:    lands,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  ainokWayfarerPick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

// ainokWayfarerPick is the land taken, or the counter. An on-disk key.
var ainokWayfarerPick = askedPick("choose-cards/ainok-wayfarer-land-or-counter", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}
	return ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneHand}.Apply(ctx)
})
