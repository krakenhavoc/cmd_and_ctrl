package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pavel Chekov, First Posting — Legendary Creature — Human Pilot
// {2}{G}{U}, 1/4 (Star Trek Commander, #4974):
//
//	"If you would draw a card, you may instead choose land or nonland
//	 and reveal cards from the top of your library until you reveal a
//	 card of the chosen kind. Put that card into your hand and put all
//	 other cards revealed this way on the bottom of your library in a
//	 random order."
//
// A "may" draw replacement (CR 614.1a, 614.11): each draw of yours,
// a draw step's or an effect's, asks first whether to replace it; "no"
// draws as usual. "Yes" cancels the draw (nothing is drawn, so a "whenever
// you draw" trigger stays quiet) and asks land or nonland, then reveals
// from the top until a card of that kind, which goes to your hand; every
// other card revealed goes to the bottom in a random order. A library
// with no such card is revealed whole and all of it goes under; an empty
// library reveals nothing and nobody loses, because nothing was drawn
// (CR 704.5b). A "draw three" asks three times, one after the other
// (CR 614.11a). The whole run waits on restore points as data: the
// parked draw and the revealed cards ride the prompt and the tail.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "473aba5a-5c87-4294-a556-f8aca0e5fe39",
		Name:         "Pavel Chekov, First Posting",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:  []game.EventKind{game.EventDrawCard},
			Optional: true,
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventDraw && ev.DrawCount > 0 && src != nil && ev.DrawPlayer == src.Controller
			},
			DrawInstead: game.RegisterDrawInstead("pavel-chekov-first-posting", pavelChekovChooseKind),
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			PromptQuestion: "Pavel Chekov, First Posting — instead of drawing, choose land or nonland and reveal until you reveal a card of that kind?",
			Label:          "Pavel Chekov, First Posting: choose land or nonland and reveal until one instead of drawing",
		}},
	})
}

const (
	pavelChekovLabel   = "Pavel Chekov, First Posting"
	pavelChekovLand    = "Land"
	pavelChekovNonland = "Nonland"
)

// pavelChekovChooseKind is the replacement's body: the land-or-nonland
// question, with the parked draw riding the prompt.
func pavelChekovChooseKind(g *game.Game, drawer, source uuid.UUID, done game.DrawDone) error {
	g.QueueOptionPickForEffect(game.OptionPickPrompt{
		Chooser:  drawer,
		Source:   source,
		Question: pavelChekovLabel + " — choose land or nonland",
		Options:  []game.ChoiceOption{{Label: pavelChekovLand}, {Label: pavelChekovNonland}},
		ThenKey:  pavelChekovKindChosen,
		Carry:    []uuid.UUID{done.Park(g)},
	})
	return nil
}

// pavelChekovKindChosen reveals until a card of the chosen kind and takes
// it. An on-disk key.
var pavelChekovKindChosen = game.RegisterOptionPickThen("option-pick/pavel-chekov-first-posting-kind", func(g *game.Game, r game.OptionPicked) error {
	if len(r.Carry) == 0 {
		return nil
	}
	if r.Option == nil {
		return game.DrawDoneByID(r.Carry[0]).Run(g)
	}
	wantLand := r.Option.Label == pavelChekovLand
	p := g.PlayerByIDForEffect(r.Chooser)
	var run []uuid.UUID
	hit := uuid.Nil
	if p != nil && p.Library != nil {
		for i := len(p.Library.Cards) - 1; i >= 0; i-- {
			c := p.Library.Cards[i]
			run = append(run, c.InstanceID)
			if !c.IsToken() && c.IsLand() == wantLand {
				hit = c.InstanceID
				break
			}
		}
	}
	if len(run) > 0 {
		g.RevealForEffect(game.RevealSpec{
			Player: r.Chooser,
			Source: r.Source,
			Reason: pavelChekovLabel + " — reveal until a " + map[bool]string{true: "land", false: "nonland"}[wantLand] + " card",
			Cards:  run,
		})
	}
	var picked []uuid.UUID
	if hit != uuid.Nil {
		picked = []uuid.UUID{hit}
	}
	tail := game.KeyedTail{Ref: pavelChekovTaken, Player: r.Chooser, Source: r.Source,
		Carry: append([]uuid.UUID{r.Carry[0]}, run...)}
	return finishTake(g, r.Source, r.Chooser, false, pavelChekovLabel+" — put the revealed card into your hand", tail, picked)
})

// pavelChekovTaken is the rest once the card has reached the hand: every
// other revealed card (Carry[1:]) goes to the bottom in a random order,
// then the parked draw (Carry[0]) is paid. An on-disk key.
var pavelChekovTaken = game.RegisterTail("pavel-chekov-first-posting/taken", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	res := takeResult(g, o.Tail.Source, o.Tail.Player, o.Tail.Carry[1:], o.IDs)
	if err := TakeRestOnBottomInRandomOrder(g, res); err != nil {
		return takeThenErr(err)
	}
	return takeThenErr(game.DrawDoneByID(o.Tail.Carry[0]).Run(g))
})
