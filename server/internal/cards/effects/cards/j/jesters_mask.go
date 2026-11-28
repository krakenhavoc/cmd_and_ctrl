package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jester's Mask — Artifact {5} (Ice Age, S82 OLD batch 846, #5636):
//
//	"This artifact enters tapped.
//	 {1}, {T}, Sacrifice this artifact: Target opponent puts the cards from
//	 their hand on top of their library. Search that player's library for that
//	 many cards. That player puts those cards into their hand, then shuffles."
//
// The count is fixed as the ability resolves: how many cards the opponent's
// hand held. They put the whole hand on top in the order they choose
// (PutInLibraryInChosenOrderThenForEffect, the Brainstorm prompt; a hand of one
// has no order to ask), and the rest of the ability is its keyed tail. The
// search is by the Mask's controller (Bribery's LibraryOwner): they choose any
// that many cards of the opponent's library, which reach that player's hand as
// the owner's, and the library is shuffled. An empty hand searches nothing and
// the library is still shuffled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bcc797b9-d2e5-4fde-8d80-c890c871f344",
		Name:         "Jester's Mask",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: Target opponent puts the cards from their hand on top of their library. Search that player's library for that many cards. That player puts those cards into their hand, then shuffles.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				opp, ok := effects.FirstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				hand := effects.AllHandCardIDs(g, opp)
				then := ctx.ItemTail(jestersMaskSearch)
				then.Player = ctx.Controller()
				then.N = len(hand)
				then.Carry = []uuid.UUID{opp}
				return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
					Chooser:   opp,
					Source:    ctx.Source(),
					Cards:     hand,
					From:      game.ZoneHand,
					Placement: game.LibraryPlaceTop,
					Reason:    "Jester's Mask — put your hand on top of your library, in what order? (the first is on top)",
					ThenTail:  then,
				})
			},
		}},
	})
}

// jestersMaskSearch is the rest of the ability: the Mask's controller
// searches the opponent's library for as many cards as the hand held, and
// the opponent shuffles. An on-disk key: never renamed, never reused.
var jestersMaskSearch = effects.CardTail("jesters-mask/search-for-that-many", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	opp := o.Tail.Carry[0]
	if o.Tail.N <= 0 {
		return effects.ShuffleLibrary{Player: opp}.Apply(ctx)
	}
	return effects.SearchLibrary{
		Player:       o.Tail.Player,
		LibraryOwner: opp,
		Dest:         game.ZoneHand,
		Limit:        o.Tail.N,
		Shuffle:      true,
		Reason:       "Jester's Mask — search that player's library for that many cards",
		Source:       ctx.Source(),
	}.Apply(ctx)
})
