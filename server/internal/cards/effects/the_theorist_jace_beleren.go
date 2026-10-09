package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Theorist, Jace Beleren — Legendary Planeswalker — Jace {2}{U}{U},
// loyalty 3 (Reality Fracture, tracker #2795):
//
//		"At the beginning of each opponent's draw step, you draw a card.
//		 +1: Create a 1/1 blue Illusion creature token.
//		 −2: For each opponent, return up to one target artifact or creature
//		     that player controls to its owner's hand.
//		 −6: Draw three cards. Then put X +1/+1 counters on each creature you
//		     control, where X is the number of cards in your hand."
//
//	  - The static trigger is Font of Mythos' shape: one per opponent's draw
//	    step, the drawer being the event's actor.
//	  - "For each opponent, up to one target ... that player controls" is one
//	    "any number" clause with the set rule EachDifferentController (the
//	    Windgrace's Judgment reading): every pick is an opponent's artifact or
//	    creature and no two share a controller, which is "one per opponent".
//	  - The −6 draws first and counts the hand once the draw is finished
//	    (DrawNThenForEffect), so a draw that paused on a prompt is counted
//	    after it, and X is fixed once for every creature (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "89a6e876-0b00-4671-9652-766fd7ef9bf2",
		Name:         "The Theorist, Jace Beleren",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginDrawStep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil && ev.Actor != source.Controller
			},
			Key: "The Theorist, Jace Beleren — you draw a card",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Create a 1/1 blue Illusion creature token.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("1/1 blue Illusion"),
						N:          1,
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "−2: For each opponent, return up to one target artifact or creature that player controls to its owner's hand.",
				Cost:  LoyaltyCost(-2),
				Targets: TargetPermanent("up to one target artifact or creature each opponent controls",
					And(Or(Artifact(), Creature()), OpponentControls())).
					WithCount(0, 0).EachDifferent(EachDifferentController()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return b35BounceChosenPermanents(NewContext(g, item))
				},
			},
			{
				Label: "−6: Draw three cards. Then put X +1/+1 counters on each creature you control, where X is the number of cards in your hand.",
				Cost:  LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return g.DrawNThenForEffect(item.Controller, 3, game.DrawThen{
						Ref: theoristCountersFromHand, Player: item.Controller, Source: item.SourceCardID,
					})
				},
			},
		},
	})
}

// theoristCountersFromHand is the −6's second sentence: X is the size of
// the player's hand now, and every creature they control gets X +1/+1
// counters.
var theoristCountersFromHand = game.RegisterDrawThen("the-theorist/counters-equal-to-hand", func(g *game.Game, d game.DrawThen) error {
	p := g.PlayerByIDForEffect(d.Player)
	if p == nil || p.Hand == nil {
		return nil
	}
	x := p.Hand.Size()
	if x <= 0 {
		return nil
	}
	g.RecomputeLayersIfStaleLocked()
	var creatures []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == d.Player && c.IsCreature() {
			creatures = append(creatures, c.InstanceID)
		}
	}
	for _, id := range creatures {
		if err := nothingIfGone(g.AddCounterForEffect(id, game.CounterPlusOne, x)); err != nil {
			return err
		}
	}
	return nil
})
