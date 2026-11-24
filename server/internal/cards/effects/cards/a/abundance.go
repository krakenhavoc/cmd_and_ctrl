package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Abundance — Enchantment {2}{G}{G} (S82 GAP2, #5948, #6089):
//
//	"If you would draw a card, you may instead choose land or nonland and
//	 reveal cards from the top of your library until you reveal a card of
//	 the chosen kind. Put that card into your hand and put all other cards
//	 revealed this way on the bottom of your library in any order."
//
// A "may" draw replacement (CR 614.1a, 614.11): each draw of yours, a draw
// step's or an effect's, asks first whether to replace it; "no" draws as
// usual. "Yes" cancels the draw (nothing is drawn, so a "whenever you draw"
// trigger stays quiet) and asks land or nonland, then reveals from the top
// until a card of that kind, which goes to your hand; every other card
// revealed goes to the bottom in the order you choose (the put_in_library
// prompt). A library with no such card is revealed whole and all of it goes
// under; an empty library reveals nothing and nobody loses, because nothing
// was drawn (CR 614.11, 704.5b). A "draw three" asks three times, one after
// the other (CR 614.11a). Pavel Chekov, First Posting's shape, with the
// player's order instead of a random one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eccc9a54-2b56-4a02-9926-258d5b2e25fb",
		Name:         "Abundance",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:  []game.EventKind{game.EventDrawCard},
			Optional: true,
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventDraw && ev.DrawCount > 0 && src != nil && ev.DrawPlayer == src.Controller
			},
			DrawInstead: game.RegisterDrawInstead("abundance", abundanceChooseKind),
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			PromptQuestion: "Abundance — instead of drawing, choose land or nonland and reveal until you reveal a card of that kind?",
			Label:          "Abundance: choose land or nonland and reveal until one instead of drawing",
		}},
	})
}

const (
	abundanceLabel   = "Abundance"
	abundanceLand    = "Land"
	abundanceNonland = "Nonland"
)

// abundanceChooseKind is the replacement's body: the land-or-nonland
// question, with the parked draw riding the prompt.
func abundanceChooseKind(g *game.Game, drawer, source uuid.UUID, done game.DrawDone) error {
	g.QueueOptionPickForEffect(game.OptionPickPrompt{
		Chooser:  drawer,
		Source:   source,
		Question: abundanceLabel + " — choose land or nonland",
		Options:  []game.ChoiceOption{{Label: abundanceLand}, {Label: abundanceNonland}},
		ThenKey:  abundanceKindChosen,
		Carry:    []uuid.UUID{done.Park(g)},
	})
	return nil
}

// abundanceKindChosen reveals until a card of the chosen kind and takes it.
// An on-disk key.
var abundanceKindChosen = game.RegisterOptionPickThen("option-pick/abundance-kind", func(g *game.Game, r game.OptionPicked) error {
	if len(r.Carry) == 0 {
		return nil
	}
	if r.Option == nil {
		return game.DrawDoneByID(r.Carry[0]).Run(g)
	}
	wantLand := r.Option.Label == abundanceLand
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
		kind := "nonland"
		if wantLand {
			kind = "land"
		}
		g.RevealForEffect(game.RevealSpec{
			Player: r.Chooser,
			Source: r.Source,
			Reason: abundanceLabel + " — reveal until a " + kind + " card",
			Cards:  run,
		})
	}
	var picked []uuid.UUID
	if hit != uuid.Nil {
		picked = []uuid.UUID{hit}
	}
	tail := game.KeyedTail{Ref: abundanceTaken, Player: r.Chooser, Source: r.Source,
		Carry: append([]uuid.UUID{r.Carry[0]}, run...)}
	return effects.FinishTake(g, r.Source, r.Chooser, false, abundanceLabel+" — put the revealed card into your hand", tail, picked)
})

// abundanceTaken is the rest once the card has reached the hand: every
// other revealed card (Carry[1:]) goes to the bottom in the order the
// player chooses, then the parked draw (Carry[0]) is paid. An on-disk key.
var abundanceTaken = game.RegisterTail("abundance/taken", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	res := effects.TakeResult(g, o.Tail.Source, o.Tail.Player, o.Tail.Carry[1:], o.IDs)
	return effects.TakeThenErr(g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   res.Player,
		Source:    res.Source,
		Cards:     res.Rest,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceBottom,
		Reason:    effects.LibraryOrderLabel(g, res.Source, "put the other revealed cards on the bottom of your library in any order"),
		ThenTail:  game.KeyedTail{Ref: abundanceOrdered, Player: res.Player, Source: res.Source, Carry: []uuid.UUID{o.Tail.Carry[0]}},
	}))
})

// abundanceOrdered pays the parked draw once the other cards are under.
// An on-disk key.
var abundanceOrdered = game.RegisterTail("abundance/ordered", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return game.DrawDoneByID(o.Tail.Carry[0]).Run(g)
})
