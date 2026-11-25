package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Priest of the Wakening Sun — Creature — Human Cleric {W}, 1/1 (Ixalan):
//
//	"At the beginning of your upkeep, you may reveal a Dinosaur card from your
//	 hand. If you do, you gain 2 life.
//	 {3}{W}{W}, Sacrifice this creature: Search your library for a Dinosaur card,
//	 reveal it, put it into your hand, then shuffle."
//
// The upkeep trigger is Invasion of the Giants's reveal: the Dinosaur cards in
// the hand are offered (none, no question) and the revealed one is shown to the
// table; the 2 life follows only a reveal. The sacrifice is a cost, so the Priest
// is gone whatever the search finds, and the search is Spinewoods Armadillo's:
// it may find nothing (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc11efa2-27ca-4745-8178-31576539512a",
		Name:         "Priest of the Wakening Sun",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Priest of the Wakening Sun — you may reveal a Dinosaur card from your hand; if you do, you gain 2 life", priestOfTheWakeningSunReveal),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}{W}, Sacrifice this creature: Search your library for a Dinosaur card, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Tutors: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}{W}{W}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				dinosaur := ctx.TextSubtype("Dinosaur")
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.HasSubtype(dinosaur) },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Priest of the Wakening Sun — a Dinosaur card",
				}.Apply(ctx)
			},
		}},
	})
}

// priestOfTheWakeningSunReveal offers the Dinosaur cards in the controller's hand.
func priestOfTheWakeningSunReveal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil {
		return nil
	}
	dinosaur := ctx.TextSubtype("Dinosaur")
	var dinos []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.HasSubtype(dinosaur) {
			dinos = append(dinos, c.InstanceID)
		}
	}
	if len(dinos) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   ctx.Source(),
		Question: "Priest of the Wakening Sun — you may reveal a Dinosaur card from your hand (then you gain 2 life)",
		Cards:    dinos,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneHand,
		ThenKey:  priestOfTheWakeningSunRevealed,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// priestOfTheWakeningSunRevealed reveals the chosen Dinosaur and gains the life.
// An on-disk key: never renamed, never reused.
var priestOfTheWakeningSunRevealed = effects.ItemChoice("choose-cards/priest-of-the-wakening-sun-reveal", func(ctx *effects.Context, r game.CardsChosen) error {
	if !r.Answered || len(r.Picked) == 0 {
		return nil
	}
	if err := (effects.RevealCards{Cards: r.Picked[:1], Reason: "Priest of the Wakening Sun — reveal a Dinosaur card"}).Apply(ctx); err != nil {
		return err
	}
	return effects.GainLife{Amount: 2}.Apply(ctx)
})
