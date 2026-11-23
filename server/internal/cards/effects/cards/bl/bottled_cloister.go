package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bottled Cloister — Artifact {4} (Ravnica Remastered, #5948):
//
//	"At the beginning of each opponent's upkeep, exile all cards from your
//	 hand face down.
//	 At the beginning of your upkeep, return all cards you own exiled with
//	 this artifact to your hand, then draw a card."
//
// The exile is Memory Jar's face-down hand exile, one batch, with each card
// linked to this artifact object (CR 607.2a; Game.ExileHandFaceDownLinkedKeyedForEffect),
// so nobody may look at them (CR 406.3). The return reads those links back
// (ExiledWithForEffect) and takes the ones the controller owns; a card the
// artifact exiled from an opponent's hand cannot exist, since only its
// controller's hand is exiled. The draw follows the return (CR 608.2c) and
// happens even when nothing came back. If the artifact leaves and returns
// it is a new object and the old exiles are lost with it (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1fd5e5e7-f23d-4891-86c4-268195ea09fc",
		Name:         "Bottled Cloister",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachOpponentsUpkeep("Bottled Cloister — exile all cards from your hand face down", bottledCloisterExile),
			effects.AtYourUpkeep("Bottled Cloister — return all cards you own exiled with this artifact to your hand, then draw a card", bottledCloisterReturn),
		},
	})
}

func bottledCloisterExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	return g.ExileHandFaceDownLinkedKeyedForEffect(ref, ctx.Controller(), game.KeyedTail{})
}

func bottledCloisterReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.Owner == ctx.Controller() {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
	}
	return g.BounceCardsToHandKeyedForEffect(ids, ctx.Tail(bottledCloisterDraw))
}

var bottledCloisterDraw = effects.CardTail("tail/bottled-cloister-draw", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
})
