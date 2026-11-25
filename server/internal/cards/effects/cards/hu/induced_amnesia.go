package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Induced Amnesia — Enchantment {2}{U}:
//
//	"When this enchantment enters, target player exiles all cards from their
//	 hand face down, then draws that many cards.
//	 When this enchantment is put into a graveyard from the battlefield, return
//	 the exiled cards to their owner's hand."
//
// Bottled Cloister's linked exile: the hand goes to exile face down in one
// batch (CR 406.3), every card linked to THIS object (CR 607.2a,
// ExileHandFaceDownLinkedKeyedForEffect), and the tail draws as many cards as
// landed in exile. The second ability reads the cards exiled with the
// enchantment as it last existed (CR 608.2h), so it still finds them after the
// enchantment has gone, and it sends each to its owner's hand. Leaving the
// battlefield any other way (exile, bounce) leaves the cards exiled: only a
// graveyard returns them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f01c124-0518-428c-9efe-7f3741f3c109",
		Name:         "Induced Amnesia",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Induced Amnesia — target player exiles all cards from their hand face down, then draws that many cards",
				inducedAmnesiaExile), effects.TargetPlayer("target player")),
			effects.WhenThisDies("Induced Amnesia — return the exiled cards to their owner's hand", inducedAmnesiaReturn),
		},
	})
}

func inducedAmnesiaExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := effects.TargetedPlayerOf(ctx)
	ref, ok := ctx.SourceRef()
	if player == uuid.Nil || !ok {
		return nil
	}
	return g.ExileHandFaceDownLinkedKeyedForEffect(ref, player, ctx.ItemTail(inducedAmnesiaDraw))
}

// inducedAmnesiaDraw draws as many cards as the hand's exile landed. An
// on-disk key: never renamed, never reused.
var inducedAmnesiaDraw = effects.CardTail("tail/induced-amnesia-draw", func(ctx *effects.Context, o game.TailOutcome) error {
	player := effects.TargetedPlayerOf(ctx)
	if player == uuid.Nil || len(o.IDs) == 0 {
		return nil
	}
	return effects.DrawCards{Player: player, N: len(o.IDs)}.Apply(ctx)
})

func inducedAmnesiaReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
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
	return g.BounceCardsToHandKeyedForEffect(ids, game.KeyedTail{})
}
