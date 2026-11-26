package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Steam Augury — Instant {2}{U}{R} (Commander 2015, S82 OLD #5546):
//
//	"Reveal the top five cards of your library and separate them into two
//	 piles. An opponent chooses one of those piles. Put that pile into your
//	 hand and the other into your graveyard."
//
// Intrude on the Mind without the Thopter: you separate the five revealed
// cards and the opponent you name (the ruling: you decide which opponent
// chooses) takes a pile for your hand; either pile may be empty. The hand move
// is a keyed bounce (a commander asks its question there), so the graveyard
// half is its tail, told which cards to bin.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0aa556a6-66aa-42f1-ba41-0001e10c20a4",
		Name:         "Steam Augury",
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			revealed := ctx.Game.RevealTopOfLibraryForEffect(ctx.Controller(), ctx.Source(), 5,
				"Steam Augury — reveal the top five cards of your library")
			if len(revealed) == 0 {
				return nil
			}
			return effects.ChoosePlayer{
				Among:    effects.Opponents,
				Question: "Steam Augury — choose an opponent to choose one of the two piles",
				ThenKey:  steamAuguryChooser,
				Carry:    revealed,
			}.Apply(ctx)
		},
	})
}

var steamAuguryChooser = effects.ChoosePlayerThen("option-pick/steam-augury-chooser", func(ctx *effects.Context, chooser uuid.UUID, revealed []uuid.UUID) error {
	if chooser == uuid.Nil {
		// No opponent is left to choose (CR 800.4a): everything is one pile
		// and nobody takes it, so the cards go to the graveyard.
		return steamAugurySettle(ctx, nil, revealed)
	}
	return effects.PileSplit{
		Splitter:      ctx.Controller(),
		Chooser:       chooser,
		Owner:         ctx.Controller(),
		SplitQuestion: "Steam Augury — separate these five cards into two piles",
		PickQuestion:  "Steam Augury — choose a pile to put into their hand; the other goes to their graveyard",
		Cards:         revealed,
		ThenTail:      ctx.Tail(steamAuguryTake),
	}.Apply(ctx)
})

var steamAuguryTake = effects.PileTail("then/steam-augury-take", func(ctx *effects.Context, taken, left []uuid.UUID, _ game.TailOutcome) error {
	return steamAugurySettle(ctx, taken, left)
})

func steamAugurySettle(ctx *effects.Context, toHand, toGraveyard []uuid.UUID) error {
	rest := effects.WithCarry(ctx.Tail(steamAuguryRest), toGraveyard...)
	if len(toHand) == 0 {
		return rest.Run(ctx.Game, nil)
	}
	return ctx.Game.BounceCardsToHandKeyedForEffect(toHand, rest)
}

// steamAuguryRest is "the other into your graveyard".
var steamAuguryRest = effects.CardTail("tail/steam-augury-rest", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.Tail.Carry {
		if ctx.Game.FindCardZoneForEffect(id) == nil {
			continue
		}
		if err := ctx.Game.PutIntoGraveyardForEffect(id); err != nil {
			return err
		}
	}
	return nil
})
