package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bomat Courier — Artifact Creature — Construct {1}, 1/1 (Kaladesh, #5539):
//
//	"Haste
//	 Whenever this creature attacks, exile the top card of your library face
//	 down. (You can't look at it.)
//	 {R}, Discard your hand, Sacrifice this creature: Put all cards exiled with
//	 this creature into their owners' hands."
//
// The attack trigger is Expert-Level Safe's face-down exile: the top card goes
// to exile with nobody able to look at it (CR 406.3a), linked to this creature
// as the object it is now (CR 607.2a); a creature that has already left exiles
// it with no link. The activated ability pays its whole cost as it is
// activated (CR 602.2b), the hand discarded and the Courier sacrificed, then
// finds the cards by the object it was, from last-known information (CR
// 608.2h), and puts every one into its owner's hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a88f3881-4f43-4473-965f-f5c24ae5752f",
		Name:            "Bomat Courier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Bomat Courier — exile the top card of your library face down", bomatCourierExile),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Discard your hand, Sacrifice this creature: Put all cards exiled with this creature into their owners' hands.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.DiscardYourHand(), effects.SacrificeThis()),
			Effect:  bomatCourierReturn,
		}},
	})
}

// bomatCourierExile exiles the top card of its controller's library face down,
// linked to the Courier.
func bomatCourierExile(g *game.Game, item *game.StackItem) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Library == nil || len(p.Library.Cards) == 0 {
		return nil
	}
	top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
	_, err := g.ExileFaceDownLinkedForEffect(item.SourceCardID, item.Controller, top)
	return err
}

// bomatCourierReturn puts every card exiled with the Courier into its owner's
// hand.
func bomatCourierReturn(g *game.Game, item *game.StackItem) error {
	ref, ok := effects.NewContext(g, item).SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		ids = append(ids, c.InstanceID)
	}
	if len(ids) == 0 {
		return nil
	}
	g.BounceCardsToHandForEffect(ids)
	return nil
}
