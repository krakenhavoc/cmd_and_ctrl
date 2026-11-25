package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scroll Rack — Artifact {2}:
//
//	"{1}, {T}: Exile any number of cards from your hand face down. Put
//	 that many cards from the top of your library into your hand. Then
//	 look at the exiled cards and put them on top of your library in any
//	 order."
//
// Composed from the engine's existing doors (#6165):
//
//   - The controller picks any number of cards from their hand, none
//     included (a choose_cards prompt over the hand, LiveZone hand).
//   - Those cards are exiled face down, unseen (CR 406.3: no player may
//     examine them, the opponents above all).
//   - "That many" is the number that reached exile; that many cards are
//     taken from the top of the library into the hand. It is a put, not a
//     draw (CR 121.1): nothing that watches draws sees it.
//   - The controller looks at the exiled cards and orders them on top of
//     their library (the put_in_library prompt, From exile; the chooser
//     alone is made a knower, ADR 0088).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     ScrollRackOracleID,
		Name:         "Scroll Rack",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{1}, {T}: Exile any number of cards from your hand face down. Put that many cards from the top of your library into your hand. Then look at the exiled cards and put them on top of your library in any order.",
			Cost:  effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			// ADR 0142: card selection answers nothing.
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				hand := effects.AllHandCardIDs(g, item.Controller)
				if len(hand) == 0 {
					return nil
				}
				g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:  item.Controller,
					Source:   item.SourceCardID,
					Question: "Scroll Rack — exile any number of cards from your hand face down",
					Cards:    hand,
					Min:      0,
					Max:      len(hand),
					LiveZone: game.ZoneHand,
					ThenKey:  scrollRackExile,
					Asked:    game.NewAskedItem(item),
				})
				return nil
			},
		}},
	})
}

// ScrollRackOracleID is the card's oracle_id.
const ScrollRackOracleID = "648bd03d-247e-4d6f-87a9-12a67688cbf5"

// scrollRackExile exiles the picked cards face down. An on-disk key.
var scrollRackExile = effects.AskedPick("choose-cards/scroll-rack-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	tail := ctx.Tail(scrollRackTake)
	return ctx.Game.ExileCardsFaceDownChosenKeyedForEffect(ctx.Controller(), r.Picked, tail)
})

// scrollRackTake puts as many cards from the top of the library into the
// hand as reached exile, then has the exiled cards put back on top. An
// on-disk key.
var scrollRackTake = game.RegisterTail("scroll-rack/take", func(g *game.Game, o game.TailOutcome) error {
	exiled := o.IDs
	back := game.KeyedTail{Ref: scrollRackPutBack, Player: o.Tail.Player, Source: o.Tail.Source, Carry: append([]uuid.UUID(nil), exiled...)}
	if len(exiled) == 0 {
		return nil
	}
	p := g.PlayerByIDForEffect(o.Tail.Player)
	var top []uuid.UUID
	if p != nil && p.Library != nil {
		for i := len(p.Library.Cards) - 1; i >= 0 && len(top) < len(exiled); i-- {
			top = append(top, p.Library.Cards[i].InstanceID)
		}
	}
	if len(top) == 0 {
		return back.Run(g, nil)
	}
	return g.TakeFromLibraryToHandKeyedForEffect(o.Tail.Player, top, back)
})

// scrollRackPutBack orders the exiled cards (the Carry) on top of the
// library. An on-disk key.
var scrollRackPutBack = game.RegisterTail("scroll-rack/put-back", func(g *game.Game, o game.TailOutcome) error {
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   o.Tail.Player,
		Source:    o.Tail.Source,
		Cards:     o.Tail.Carry,
		From:      game.ZoneExile,
		Placement: game.LibraryPlaceTop,
		Reason:    "Scroll Rack — put the exiled cards on top of your library in any order",
	})
})
