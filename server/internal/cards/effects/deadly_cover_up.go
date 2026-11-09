package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deadly Cover-Up — Sorcery {3}{B}{B} (Murders at Karlov Manor, #4469):
//
//	"As an additional cost to cast this spell, you may collect evidence 6.
//	 Destroy all creatures. If evidence was collected, exile a card from an
//	 opponent's graveyard. Then search its owner's graveyard, hand, and
//	 library for any number of cards with that name and exile them. That
//	 player shuffles, then draws a card for each card exiled from their hand
//	 this way."
//
// The optional cost is OptionalCollectEvidence (#5193); "if evidence was
// collected" reads the announcement (ctx.EvidenceCollected). The sweep goes
// first and its keyed tail starts the second half, so a commander caught in
// it asks its question before anything else happens. The exile is the
// caster's choice among the cards in ANY opponent's graveyard (a basic land
// included, the ruling), read before it moves: the name, as a card in a
// graveyard has it, is searched for in its owner's graveyard, hand and
// library (the extraction family's search, extractionNames; "any number",
// so each match is the caster's to take or leave). That player then
// shuffles, and draws once for each card the search took out of their hand,
// counted from the moves themselves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "04e2afbe-0c08-42bc-93e6-662066a958bf",
		Name:          "Deadly Cover-Up",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{OptionalCollectEvidence(6)},
		Purpose:       game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			// "If evidence was collected" is read from the announcement now and
			// carried: the tail's copy of the item does not hold what was paid.
			tail := ctx.ItemTail(deadlyCoverUpAfterSweep)
			if ctx.EvidenceCollected() {
				tail.N = 1
			}
			return DestroyAllMatching{Match: Creature(), ThenTail: tail}.Apply(ctx)
		},
	})
}

// deadlyCoverUpAfterSweep is "if evidence was collected, exile a card from an
// opponent's graveyard": the pick over every opponent's graveyard.
var deadlyCoverUpAfterSweep = CardTail("tail/deadly-cover-up-after-sweep", func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.N == 0 {
		return nil
	}
	var cards []uuid.UUID
	for _, opp := range ctx.Opponents() {
		cards = append(cards, graveyardCardIDs(ctx, opp, func(game.Card) bool { return true })...)
	}
	if len(cards) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Deadly Cover-Up — exile a card from an opponent's graveyard",
		Cards:    cards,
		Min:      1,
		Max:      1,
		Asked:    game.NewAskedItem(ctx.Item),
		ThenKey:  deadlyCoverUpPicked,
	})
	return nil
})

// deadlyCoverUpPicked is the pick answered: the card is exiled and its name
// carried to the search, with its owner.
var deadlyCoverUpPicked = game.RegisterChooseCardsThen("choose-cards/deadly-cover-up-exile", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) != 1 || r.Asked == nil {
		return nil
	}
	ctx := NewContext(g, r.Asked)
	c, ok := g.LookupCardForEffect(r.Picked[0])
	if z := g.FindCardZoneForEffect(r.Picked[0]); !ok || z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	names := strings.Join(extractionNames(c), "\n")
	return ExileTarget{
		Target:   r.Picked[0],
		ThenTail: withWord(withCarry(ctx.ItemTail(deadlyCoverUpExiled), c.Owner), names),
	}.Apply(ctx)
})

// deadlyCoverUpExiled is "then search its owner's graveyard, hand, and
// library for any number of cards with that name and exile them".
var deadlyCoverUpExiled = CardTail("tail/deadly-cover-up-exiled", func(ctx *Context, o game.TailOutcome) error {
	victim := tailCarry0(o)
	if victim == uuid.Nil {
		return nil
	}
	names := strings.Split(o.Tail.Word, "\n")
	if len(names) == 1 && names[0] == "" {
		names = nil
	}
	return SearchLibrary{
		Player:               ctx.Controller(),
		LibraryOwner:         victim,
		Predicate:            sharesANameWith(names),
		Dest:                 game.ZoneExile,
		Unbounded:            true,
		AlsoGraveyardAndHand: true,
		Shuffle:              true,
		Reason:               "Deadly Cover-Up — search their graveyard, hand and library for any number of cards with that name",
		ThenTail:             withCarry(ctx.ItemTail(deadlyCoverUpDraw), victim),
	}.Apply(ctx)
})

// deadlyCoverUpDraw is "that player draws a card for each card exiled from
// their hand this way", counted from the zone moves the search made.
var deadlyCoverUpDraw = CardTail("tail/deadly-cover-up-draw", func(ctx *Context, o game.TailOutcome) error {
	victim := tailCarry0(o)
	if victim == uuid.Nil || len(o.IDs) == 0 {
		return nil
	}
	found := map[uuid.UUID]bool{}
	for _, id := range o.IDs {
		found[id] = true
	}
	n := 0
	for _, ev := range ctx.Game.EventsThisTurn() {
		if ev.Kind == game.EventZoneMove && found[ev.CardID] && ev.OldZone == game.ZoneHand && ev.NewZone == game.ZoneExile {
			n++
			delete(found, ev.CardID)
		}
	}
	if n == 0 {
		return nil
	}
	return DrawCards{Player: victim, N: n}.Apply(ctx)
})
