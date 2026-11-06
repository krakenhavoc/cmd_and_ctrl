package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ostrich-Horse — Creature — Bird Horse {2}{G}, 3/1:
//
//	"When this creature enters, mill three cards. You may put a land
//	 card from among them into your hand. If you don't, put a +1/+1
//	 counter on this creature."
//
// Six's "mill, then you may take one from among them" (mill_then_take.go):
// the candidates are the cards this mill put into the graveyard that are
// still there. "If you don't" covers both a decline and a mill that turned
// up no land, so with no candidate the counter goes on at once, with no
// prompt. The counter goes on this creature only while it is still on the
// battlefield as the same object (CounterOnThis).
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "2b6cb199-1564-4214-b24f-624dc962c907",
		Name:         "Ostrich-Horse",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ostrich-Horse — mill three cards; you may put a land card from among them into your hand, or put a +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{N: 3, ThenTail: ctx.ItemTail(ostrichHorseMilled)}.Apply(ctx)
				}),
		},
	})
}

var ostrichHorseMilled = CardTail("ostrich-horse/after-mill", func(ctx *Context, o game.TailOutcome) error {
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
		Question: "Ostrich-Horse — you may put a land card from among them into your hand (if you don't, it gets a +1/+1 counter)",
		Cards:    lands,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  ostrichHorseLandOrCounter,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

// ostrichHorseLandOrCounter is the pick's answer: the land to hand, or,
// with none picked, the +1/+1 counter.
var ostrichHorseLandOrCounter = askedPick("choose-cards/ostrich-horse-land-or-counter", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}
	return ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneHand}.Apply(ctx)
})
