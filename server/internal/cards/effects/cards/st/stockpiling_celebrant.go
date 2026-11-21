package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stockpiling Celebrant — Creature — Dwarf Knight {2}{W}, 3/2 (Wilds of
// Eldraine, #4471):
//
//	"When this creature enters, you may return another target nonland
//	 permanent you control to its owner's hand. If you do, scry 2."
//
// The target is chosen as the trigger goes on the stack, and the "you
// may" is asked as it resolves (Optional). The return is Mirkwood
// Nurturer's bounce with a continuation: the scry is the move's
// registered tail, run only when the permanent reached a hand
// (oneWentThisWay), so a commander that went to the command zone
// instead scries nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "171150fb-dbe1-4a07-a7ac-5f45762a26b0",
		Name:         "Stockpiling Celebrant",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Stockpiling Celebrant — you may return another nonland permanent you control to its owner's hand; if you do, scry 2", stockpilingCelebrantBounce),
				effects.Another(effects.TargetPermanent("another target nonland permanent you control", effects.Nonland(), effects.YouControl())),
			), "Stockpiling Celebrant — return that permanent to its owner's hand and scry 2?"),
		},
	})
}

func stockpilingCelebrantBounce(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return effects.BounceToHand{Target: id, ThenTail: ctx.ItemTail(stockpilingCelebrantScry)}.Apply(ctx)
}

// stockpilingCelebrantScry is "if you do, scry 2", run once the
// permanent has reached a hand. An on-disk key.
var stockpilingCelebrantScry = effects.CardTail("stockpiling-celebrant/scry", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	return effects.Scry{N: 2}.Apply(ctx)
})
