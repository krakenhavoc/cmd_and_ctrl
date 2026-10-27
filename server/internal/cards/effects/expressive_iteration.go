package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Expressive Iteration — Sorcery {U}{R} (EDHREC rank 879):
//
//	"Look at the top three cards of your library. Put one of them into
//	 your hand, put one of them on the bottom of your library, and
//	 exile one of them. You may play the exiled card this turn."
//
// Two picks, in the printed order. The first is Impulse's take
// (TakeFromLibraryToHand, Max 1, mandatory). The second asks which of
// the two left goes on the bottom; the other is exiled with a play
// permission until end of turn (a land may be played from exile with
// a land drop, as Tersa Lightshatter's grant allows). It is a LOOK: no
// card is revealed.
//
// Fewer than three cards: the instructions are followed in order for
// the cards there are (the 2026-03-20 ruling). Two cards put one in the
// hand and the other on the bottom with nothing exiled; one card goes
// to the hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c7aecca5-2f67-4245-ab2d-e723d8b23a67",
		Name:         "Expressive Iteration",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			return TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
				Max:    1,
				Label:  "Expressive Iteration — put one of them into your hand",
				Key:    expressiveIterationTake,
			}.Apply(ctx)
		},
	})
}

// expressiveIterationBottomAndExile is the rest once a card is in hand:
// one of the remaining cards to the bottom, the other into exile.
func expressiveIterationBottomAndExile(g *game.Game, res TakeFromLibraryResult) error {
	rest := res.Rest
	switch len(rest) {
	case 0:
		return nil
	case 1:
		return expressiveIterationBottom(g, res.Player, res.Source, rest[0], uuid.Nil)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  res.Player,
		Source:   res.Source,
		Question: "Expressive Iteration — put one of them on the bottom of your library (the other is exiled, and you may play it this turn)",
		Cards:    rest,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  expressiveIterationBottomPicked,
		Carry:    rest,
	})
	return nil
}

// expressiveIterationBottomPicked is the card put on the bottom and the other
// exiled, keyed (#4253): the cards looked at ride the prompt as its carry.
var expressiveIterationBottomPicked = game.RegisterChooseCardsThen("choose-cards/expressive-iteration-bottom", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	var exile uuid.UUID
	for _, id := range cardsStillInALibrary(g, r.Carry) {
		if id != r.Picked[0] {
			exile = id
			break
		}
	}
	return expressiveIterationBottom(g, r.Chooser, r.Source, r.Picked[0], exile)
})

// expressiveIterationBottom puts one looked-at card on the bottom of its
// owner's library, then exiles `exile` (when it is not uuid.Nil) so that
// `player` may play it this turn.
//
// The exile is the ordering prompt's keyed tail (#4253 slice 10b), so a table
// waiting on the bottom's order is a restore point.
func expressiveIterationBottom(g *game.Game, player, source, card, exile uuid.UUID) error {
	var then game.KeyedTail
	if exile != uuid.Nil {
		then = game.KeyedTail{Ref: expressiveIterationExile, Player: player, Source: source, Carry: []uuid.UUID{exile}}
	}
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   player,
		Source:    source,
		Cards:     []uuid.UUID{card},
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceBottom,
		Reason:    "Expressive Iteration — put one of them on the bottom of your library",
		ThenTail:  then,
	})
}

// expressiveIterationExile is "exile the other. You may play it this turn",
// once the bottom card is placed: Tail.Carry is the card, Tail.Player who may
// play it. An on-disk key: never renamed, never reused.
var expressiveIterationExile = game.RegisterTail("then/expressive-iteration-exile", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return g.ExileCardWithPermissionForEffect(o.Tail.Carry[0], game.CastPermission{
		Player: o.Tail.Player,
		Label:  "Expressive Iteration — play it this turn",
	})
})
