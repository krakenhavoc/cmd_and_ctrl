package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Abstract Performance — Sorcery {5}{U}:
//
//	"Exile the top four cards of your library in a face-down pile, then
//	 exile the top four cards of your library in a face-up pile. An
//	 opponent chooses one of those piles. Put that pile into your
//	 graveyard. Look at the cards in the other pile. You may cast a spell
//	 from among them without paying its mana cost. Put the rest into your
//	 hand."
//
// The face-down pile is exiled so that nobody may look at it, its owner
// included (game.ExileTopFaceDownUnseenKeyedForEffect, CR 406.3); the
// face-up pile is an ordinary exile everyone sees. Each exile waits for
// the one before it on a keyed tail, so a short library exiles what it has
// and a commander's command-zone offer holds the rest.
//
// The controller chooses which opponent chooses (only when there is more
// than one), and that opponent picks a pile: the face-up pile's cards are
// on the prompt, the face-down pile shows only its size. The chosen pile
// goes to the graveyard (its face-down cards turn face up as they leave
// exile). The other pile is the controller's: one free cast permission
// over it, which also lets them look at a face-down pile (the grant makes
// its holder the viewer), is offered during the resolution (CR 608.2g) and
// spent by the one cast. Once that is answered, every card of the pile
// still in exile, the lands included, goes into the hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3e0e3676-e634-4e64-bdfd-4aaa5885461f",
		Name:         "Abstract Performance",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ctx.Game.ExileTopFaceDownUnseenKeyedForEffect(ctx.Controller(), 4, ctx.ItemTail(abstractPerformanceFaceDown))
		},
	})
}

// abstractPerformanceStillExiled is the members of ids still in exile.
func abstractPerformanceStillExiled(g *game.Game, ids []uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			out = append(out, id)
		}
	}
	return out
}

// abstractPerformanceFaceDown is the face-up exile, after the face-down
// one: the face-down pile rides the next tail. An on-disk key.
var abstractPerformanceFaceDown = CardTail("abstract-performance/after-face-down", func(ctx *Context, o game.TailOutcome) error {
	next := ctx.ItemTail(abstractPerformanceFaceUp)
	next.Carry = o.IDs
	return MillToZone{N: 4, To: game.ZoneExile, ThenTail: next}.Apply(ctx)
})

// abstractPerformanceFaceUp names the opponent who chooses. An on-disk key.
var abstractPerformanceFaceUp = CardTail("abstract-performance/after-face-up", func(ctx *Context, o game.TailOutcome) error {
	g := ctx.Game
	return ChoosePlayer{
		Among:    Opponents,
		Question: "Abstract Performance — choose an opponent to choose one of the piles",
		ThenKey:  abstractPerformanceChooser,
		Carry:    packIDLists(abstractPerformanceStillExiled(g, o.Tail.Carry), abstractPerformanceStillExiled(g, o.IDs)),
	}.Apply(ctx)
})

// abstractPerformanceChooser asks the chosen opponent for a pile. An
// on-disk key.
var abstractPerformanceChooser = ChoosePlayerThen("option-pick/abstract-performance-opponent", func(ctx *Context, opp uuid.UUID, carry []uuid.UUID) error {
	if opp == uuid.Nil {
		return nil
	}
	faceDown, faceUp := unpackIDPair(carry)
	return PickOption{
		Player:   opp,
		From:     ctx.Controller(),
		Question: "Abstract Performance — choose a pile; it goes into its owner's graveyard and they get the other",
		Options: []game.ChoiceOption{
			{Label: "The face-down pile (" + strconv.Itoa(len(faceDown)) + " cards)", Cards: faceDown},
			{Label: "The face-up pile (" + strconv.Itoa(len(faceUp)) + " cards)", Cards: faceUp},
		},
		ThenKey:  abstractPerformancePile,
		Carry:    carry,
		WithItem: true,
	}.Apply(ctx)
})

// abstractPerformancePile puts the chosen pile into the graveyard; the
// other pile rides the move's tail. An on-disk key.
var abstractPerformancePile = OptionPickThen("option-pick/abstract-performance-pile", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil {
		return nil
	}
	g := ctx.Game
	faceDown, faceUp := unpackIDPair(r.Carry)
	binned, kept := faceDown, faceUp
	if r.Index == 1 {
		binned, kept = faceUp, faceDown
	}
	next := ctx.ItemTail(abstractPerformanceKept)
	next.Carry = kept
	binned = abstractPerformanceStillExiled(g, binned)
	if len(binned) == 0 {
		return next.Run(g, nil)
	}
	return g.PutCardsIntoGraveyardKeyedForEffect(binned, next)
})

// abstractPerformanceKept offers the one free cast over the pile left,
// then puts the rest into the hand. An on-disk key.
var abstractPerformanceKept = CardTail("abstract-performance/offer-cast", func(ctx *Context, o game.TailOutcome) error {
	g, controller, source := ctx.Game, ctx.Controller(), ctx.Source()
	kept := abstractPerformanceStillExiled(g, o.Tail.Carry)
	if len(kept) == 0 {
		return nil
	}
	perm := game.CastPermission{
		Player:      controller,
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		LapseOnPass: game.LapseStaysInExile,
		Source:      source,
		SourceName:  "Abstract Performance",
		Label:       "Abstract Performance — cast it without paying its mana cost",
	}
	if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       "Abstract Performance",
		Cards:        kept,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Put them into your hand",
		Then:         abstractPerformanceRestToHand,
		Carry:        kept,
	}) {
		return nil
	}
	return gloriousPurposePutRest(g, controller, source, kept)
})

// abstractPerformanceRestToHand is "Put the rest into your hand", once the
// cast is answered: Carry is the pile. An on-disk key.
var abstractPerformanceRestToHand = game.RegisterChooseCardsThen("resolution-cast/abstract-performance-rest-to-hand",
	func(g *game.Game, r game.CardsChosen) error {
		return gloriousPurposePutRest(g, r.Chooser, r.Source, r.Carry)
	})
