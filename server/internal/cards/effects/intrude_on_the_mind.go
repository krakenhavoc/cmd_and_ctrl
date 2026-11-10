package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Intrude on the Mind — Instant {3}{U}{U} (Murders at Karlov Manor, #4469):
//
//	"Reveal the top five cards of your library and separate them into two
//	 piles. An opponent chooses one of those piles. Put that pile into your
//	 hand and the other into your graveyard. Create a 0/0 colorless Thopter
//	 artifact creature token with flying, then put a +1/+1 counter on it for
//	 each card put into your graveyard this way."
//
// Fact or Fiction turned around: YOU separate the cards and the opponent
// you name (the ruling: you decide which opponent chooses) takes a pile for
// your hand. A pile may be empty. The Thopter is made after both piles have
// moved, and its counters are placed on it, one per card that actually
// reached the graveyard (a card that left the library in response, or was
// replaced on its way, counts for nothing).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eedb1527-548a-435a-a3e1-8ad7d9d6d8a9",
		Name:         "Intrude on the Mind",
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			revealed := ctx.Game.RevealTopOfLibraryForEffect(ctx.Controller(), ctx.Source(), 5,
				"Intrude on the Mind — reveal the top five cards of your library")
			if len(revealed) == 0 {
				return intrudeThopter(ctx, 0)
			}
			return ChoosePlayer{
				Among:    Opponents,
				Question: "Intrude on the Mind — choose an opponent to choose one of the two piles",
				ThenKey:  intrudeChooser,
				Carry:    revealed,
			}.Apply(ctx)
		},
	})
}

var intrudeChooser = ChoosePlayerThen("option-pick/intrude-on-the-mind-chooser", func(ctx *Context, chooser uuid.UUID, revealed []uuid.UUID) error {
	if chooser == uuid.Nil {
		// No opponent is left to choose (CR 800.4a): everything is one pile
		// and nobody takes it, so the spell finishes with what it can do.
		return intrudeSettle(ctx, nil, revealed)
	}
	return PileSplit{
		Splitter:      ctx.Controller(),
		Chooser:       chooser,
		Owner:         ctx.Controller(),
		SplitQuestion: "Intrude on the Mind — separate these five cards into two piles",
		PickQuestion:  "Intrude on the Mind — choose a pile to put into their hand; the other goes to their graveyard",
		Cards:         revealed,
		ThenTail:      ctx.Tail(intrudeTake),
	}.Apply(ctx)
})

var intrudeTake = PileTail("then/intrude-on-the-mind-take", func(ctx *Context, taken, left []uuid.UUID, _ game.TailOutcome) error {
	return intrudeSettle(ctx, taken, left)
})

// intrudeSettle puts the chosen pile into the caster's hand and the other
// into their graveyard, then makes the Thopter. The hand move is a keyed
// bounce (a commander asks its question there), so the graveyard half and the
// Thopter are its tail, told which cards to bin.
func intrudeSettle(ctx *Context, toHand, toGraveyard []uuid.UUID) error {
	rest := withCarry(ctx.Tail(intrudeRest), toGraveyard...)
	if len(toHand) == 0 {
		return rest.Run(ctx.Game, nil)
	}
	return ctx.Game.BounceCardsToHandKeyedForEffect(toHand, rest)
}

// intrudeRest is "the other into your graveyard", then the Thopter with a
// +1/+1 counter for each card that reached the graveyard this way.
var intrudeRest = CardTail("tail/intrude-on-the-mind-rest", func(ctx *Context, o game.TailOutcome) error {
	g := ctx.Game
	n := 0
	for _, id := range o.Tail.Carry {
		if g.FindCardZoneForEffect(id) == nil {
			continue
		}
		if err := g.PutIntoGraveyardForEffect(id); err != nil {
			return err
		}
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneGraveyard {
			n++
		}
	}
	return intrudeThopter(ctx, n)
})

func intrudeThopter(ctx *Context, counters int) error {
	ids, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(TokenCard("0/0 colorless Thopter artifact with flying")), 1, game.TokenEntryOptions{})
	if err != nil || counters <= 0 {
		return err
	}
	for _, id := range ids {
		if err := ctx.Game.AddCounterForEffect(id, game.CounterPlusOne, counters); err != nil {
			return err
		}
	}
	return nil
}
