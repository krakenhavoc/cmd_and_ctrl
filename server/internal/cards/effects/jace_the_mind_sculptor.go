package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jace, the Mind Sculptor — Legendary Planeswalker — Jace {2}{U}{U},
// starting loyalty 3:
//
//	"+2: Look at the top card of target player's library. You may put
//	     that card on the bottom of that player's library.
//	 0: Draw three cards, then put two cards from your hand on top of
//	    your library in any order.
//	 −1: Return target creature to its owner's hand.
//	 −12: Exile all cards from target player's library, then that
//	      player shuffles their hand into their library."
//
// The +2 is what this file is for (#1298, ADR 0088's 2026-09-23
// amendment): a LOOK at ANOTHER player's library. LookAtLibraryThenPlace
// names the looker (Jace's controller) and the library's owner (the
// target) separately, marks only the looker a knower, and asks a
// `top_or_bottom` put_in_library of one card — "you may put that card
// on the bottom" is exactly that two-way choice, and it always asks.
// The card goes back to its OWNER's library either way. The target's
// owner learns nothing (CR 401.2, CR 701.20): the prompt's card
// reaches the looker's wire alone. Targeting yourself is legal and is
// the same prompt about your own top card.
//
// The 0 is Brainstorm's sentence — PutFromHandOnTopInAnyOrder after the
// draw: pick two, then order them (ADR 0088 Decision 1). The −1 is
// bounceChosenTarget (through BounceToHand, so a commander is offered the
// command zone). The −12 exiles the whole library as one simultaneous
// exit, then tucks the hand and shuffles, each from the previous step's
// continuation so a commander's CR 903.9 question (a commander in the
// library or the hand is offered the command zone) holds the rest of
// the ability until it is answered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7f77a84e-5a4b-4834-aefa-3cecc175ae8e",
		Name:         "Jace, the Mind Sculptor",
		Completeness: CompletenessFull,
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures and
		// the dev spawner.
		StartingLoyalty: 3,
		Activated: []ActivatedAbility{
			{
				Label: "+2: Look at the top card of target player's library. You may put that card on the " +
					"bottom of that player's library.",
				Cost:    LoyaltyCost(2),
				Targets: TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					victim, ok := firstLegalPlayerTarget(ctx)
					if !ok {
						return nil
					}
					return LookAtLibraryThenPlace{
						Owner:     victim,
						N:         1,
						Placement: game.LibraryPlaceTopOrBottom,
						Label:     "Jace, the Mind Sculptor — you may put that card on the bottom of that player's library",
					}.Apply(ctx)
				},
			},
			{
				Label: "0: Draw three cards, then put two cards from your hand on top of your library in any order.",
				Cost:  LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (DrawCards{Player: item.Controller, N: 3}).Apply(ctx); err != nil {
						return err
					}
					return PutFromHandOnTopInAnyOrder{
						Player: item.Controller,
						N:      2,
						Label:  "Jace, the Mind Sculptor — put two cards from your hand on top of your library",
					}.Apply(ctx)
				},
			},
			{
				Label:   "−1: Return target creature to its owner's hand.",
				Cost:    LoyaltyCost(-1),
				Targets: TargetCreature("target creature"),
				Effect:  bounceChosenTarget,
			},
			{
				Label: "−12: Exile all cards from target player's library, then that player shuffles their hand " +
					"into their library.",
				Cost:    LoyaltyCost(-12),
				Targets: TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					victim, ok := firstLegalPlayerTarget(NewContext(g, item))
					if !ok {
						return nil
					}
					return jaceUltimate(g, victim)
				},
			},
		},
	})
}

// jaceUltimate is the −12: the library exiled as one simultaneous exit,
// THEN the hand tucked and the library shuffled. Each step runs from the
// previous one's continuation, so a CR 903.9 prompt on any leg holds
// the rest. Captures only the victim's ID.
func jaceUltimate(g *game.Game, victim uuid.UUID) error {
	p := g.PlayerByIDForEffect(victim)
	if p == nil || p.Library == nil {
		return nil
	}
	library := make([]uuid.UUID, 0, len(p.Library.Cards))
	for _, c := range p.Library.Cards {
		library = append(library, c.InstanceID)
	}
	return g.ExileCardsThenForEffect(library, func(g *game.Game, _ []uuid.UUID) error {
		return g.TuckCardsToLibraryThenForEffect(allHandCardIDs(g, victim), game.TuckOptions{},
			func(g *game.Game, _ []uuid.UUID) error {
				return g.ShuffleLibraryForEffect(victim)
			})
	})
}
