package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Green Sun's Twilight — Sorcery {X}{G} (Phyrexia: All Will Be One):
//
//	"Reveal the top X plus one cards of your library. Choose a creature
//	 card and/or a land card from among them. Put those cards into your hand
//	 and the rest on the bottom of your library in a random order. If X is 5
//	 or more, instead put the chosen cards onto the battlefield or into your
//	 hand and the rest on the bottom of your library in a random order."
//
// The reveal is public. The choice is mandatory and is In the Presence of
// Ages' two slots (at most one creature card and at most one land card, a card
// that is both filling only one), asked for as many cards as the revealed
// ones allow. With X under 5 the picks go to the hand (RegisterTake) and the
// rest to the bottom in a random order. With X of 5 or more, "instead" swaps
// the destination: after the pick a second question names which chosen cards
// enter the battlefield (any number of them, under your control, untapped);
// each one not named goes into your hand, which is the printed per-card
// "onto the battlefield or into your hand". The hand cards move first, then
// the battlefield ones, then the rest go to the bottom.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0221330a-3e7a-40aa-9c78-85c88a4c1a53",
		Name:         "Green Sun's Twilight",
		Completeness: CompletenessFull,
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			revealed := ctx.Game.RevealTopOfLibraryForEffect(player, ctx.Source(), ctx.X()+1,
				"Green Sun's Twilight — reveal the top X plus one cards")
			if ctx.X() < 5 {
				return TakeFromLibraryToHand{
					Player: player,
					Cards:  revealed,
					Label:  "Green Sun's Twilight — choose a creature card and/or a land card to put into your hand",
					Key:    greenSunsTwilightTake,
				}.Apply(ctx)
			}
			return greenSunsTwilightBig(ctx, player, revealed)
		},
	})
}

var greenSunsTwilightSlots = []TakeSlot{
	{Label: "a creature card", Match: Creature(), Max: 1},
	{Label: "a land card", Match: Land(), Max: 1},
}

// greenSunsTwilightTake is the small mode's rule and rest. An on-disk key:
// never renamed, never reused.
var greenSunsTwilightTake = RegisterTake("choose-cards/green-suns-twilight-take", TakeDef{
	Slots: greenSunsTwilightSlots,
	Then:  TakeRestOnBottomInRandomOrder,
})

// greenSunsTwilightBig is the X >= 5 reading: pick the cards, then say which
// of them enter the battlefield.
func greenSunsTwilightBig(ctx *Context, player uuid.UUID, revealed []uuid.UUID) error {
	g := ctx.Game
	cands := libraryCardsTakeable(g, player, revealed, Or(Creature(), Land()))
	hi := newTakeSlotRule(g, player, greenSunsTwilightSlots, cands).fill(cands)
	if len(cands) == 0 || hi <= 0 {
		return greenSunsTwilightRest(g, player, revealed)
	}
	if hi >= len(cands) {
		// The only legal pick is every candidate.
		greenSunsTwilightAskPlacement(g, player, ctx.Source(), cands, revealed)
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Green Sun's Twilight — choose a creature card and/or a land card",
		Cards:    cands,
		Min:      hi,
		Max:      hi,
		Zone:     game.ZoneLibrary,
		ThenKey:  greenSunsTwilightPicked,
		Carry:    revealed,
	})
	return nil
}

// greenSunsTwilightPicked takes the chosen set and asks where it goes. An
// on-disk key: never renamed, never reused.
var greenSunsTwilightPicked = game.RegisterChooseCardsThenRule("choose-cards/green-suns-twilight-picked",
	takeRule(TakeDef{Slots: greenSunsTwilightSlots}),
	func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered || len(r.Picked) == 0 {
			return greenSunsTwilightRest(g, r.Chooser, r.Carry)
		}
		greenSunsTwilightAskPlacement(g, r.Chooser, r.Source, r.Picked, r.Carry)
		return nil
	})

func greenSunsTwilightAskPlacement(g *game.Game, player, source uuid.UUID, picked, revealed []uuid.UUID) {
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   source,
		Question: "Green Sun's Twilight — choose the cards to put onto the battlefield (the others go into your hand)",
		Cards:    picked,
		Min:      0,
		Max:      len(picked),
		Zone:     game.ZoneLibrary,
		ThenKey:  greenSunsTwilightPlaced,
		Carry:    revealed,
	})
}

// greenSunsTwilightPlaced moves the chosen cards: the named ones onto the
// battlefield, the other chosen ones into the hand, then the rest to the
// bottom in a random order. An on-disk key: never renamed, never reused.
var greenSunsTwilightPlaced = game.RegisterChooseCardsThen("choose-cards/green-suns-twilight-placed",
	func(g *game.Game, r game.CardsChosen) error {
		onField := map[uuid.UUID]bool{}
		if r.Answered {
			for _, id := range r.Picked {
				onField[id] = true
			}
		}
		var hand, field []uuid.UUID
		for _, id := range r.Candidates {
			if onField[id] {
				field = append(field, id)
			} else {
				hand = append(hand, id)
			}
		}
		if len(hand) > 0 {
			g.TakeFromLibraryToHandForEffect(r.Chooser, hand)
		}
		if len(field) > 0 {
			if _, err := g.PutCardsFromLibraryOntoBattlefieldForEffect(field, game.LibraryEntryOptions{Controller: r.Chooser}); err != nil {
				return err
			}
		}
		return greenSunsTwilightRest(g, r.Chooser, r.Carry)
	})

// greenSunsTwilightRest puts the revealed cards still in a library on the
// bottom in a random order.
func greenSunsTwilightRest(g *game.Game, player uuid.UUID, revealed []uuid.UUID) error {
	rest := libraryCardsTakeable(g, player, revealed, nil)
	if len(rest) == 0 {
		return nil
	}
	return g.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, rest)
}
