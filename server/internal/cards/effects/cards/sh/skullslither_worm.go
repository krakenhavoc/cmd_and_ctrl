package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skullslither Worm — {3}{B} Creature — Worm 3/3 (Jumpstart 2022, #5948):
//
//	"When this creature enters, each opponent discards a card. For each
//	 opponent who can't, put two +1/+1 counters on this creature."
//
// An opponent with a card in hand picks it and discards it; each
// opponent whose hand is empty (CR 701.9b) is one who can't, and the Worm
// gets two counters for each of them. The counters wait for the last
// discard (a registered run continuation, so a table waiting on a prompt is
// a restore point); the opponents who can't travel on the carry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a36dda2e-4cb8-4821-9c93-bd637ae7b333",
		Name:         "Skullslither Worm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Skullslither Worm — each opponent discards a card; two +1/+1 counters for each who can't",
				skullslitherWormEnters),
		},
	})
}

func skullslitherWormEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var asks []uuid.UUID
	carry := []uuid.UUID{item.Controller}
	for _, opp := range ctx.Opponents() {
		p := g.PlayerByIDForEffect(opp)
		if p == nil || p.Eliminated {
			continue
		}
		if p.Hand == nil || p.Hand.Size() == 0 {
			carry = append(carry, opp)
			continue
		}
		asks = append(asks, opp)
	}
	if len(asks) == 0 {
		return skullslitherWormGrow(ctx, item, len(carry)-1)
	}
	return g.PlayersDiscardThenKeyedForEffect(asks, game.DiscardPrompt{
		Source:   item.SourceCardID,
		N:        1,
		Question: "Skullslither Worm — discard a card",
		Asked:    game.NewAskedItem(item),
	}, skullslitherWormCounters, carry)
}

// skullslitherWormCounters runs once the last discard has landed: two
// +1/+1 counters for each opponent who couldn't discard (the carry after
// the controller). An on-disk key.
var skullslitherWormCounters = effects.RunThen("prompt-run/skullslither-worm-counters", func(ctx *effects.Context, r game.RunDone) error {
	if len(r.Carry) < 2 || ctx.Item == nil {
		return nil
	}
	return skullslitherWormGrow(ctx, ctx.Item, len(r.Carry)-1)
})

func skullslitherWormGrow(ctx *effects.Context, item *game.StackItem, cant int) error {
	if cant == 0 || !effects.OnBattlefield(ctx.Game, item.SourceCardID) || effects.SourceIsNewObject(ctx.Game, item) {
		return nil
	}
	return effects.AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 2 * cant}.Apply(ctx)
}
