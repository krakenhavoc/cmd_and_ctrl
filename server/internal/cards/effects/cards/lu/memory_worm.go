package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Memory Worm — Creature — Alien Worm {1}{R}, 1/1 (Doctor Who):
//
//	"Paradox — Whenever you cast a spell from anywhere other than your hand, this
//	 creature deals 2 damage to target player. That player discards a card, then
//	 draws a card. Put a +1/+1 counter on this creature."
//
// The trigger chooses its target player as it goes on the stack and re-checks it as
// it resolves. The damage is the Worm's; the damaged player chooses what to discard
// and draws after the discard has settled (an empty hand discards nothing and still
// draws); then the Worm gets its counter, if it is still the same permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8fb39f48-10a1-4d9c-870a-eeda070bf12b",
		Name:         "Memory Worm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventCast, effects.GapB88YouCastFromOutsideYourHand,
					"Paradox — 2 damage to target player; that player discards a card, then draws a card; put a +1/+1 counter on this creature",
					memoryWormParadox),
				effects.TargetPlayer("target player")),
		},
	})
}

func memoryWormParadox(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player, ok := effects.FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	if err := (effects.DealDamage{Source: item.SourceCardID, Target: player, Amount: 2}).Apply(ctx); err != nil {
		return err
	}
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   player,
		Source:   item.SourceCardID,
		N:        1,
		Question: "Memory Worm — discard a card, then draw a card",
		Asked:    game.NewAskedItem(item),
	}, memoryWormDiscarded, nil)
}

// memoryWormDiscarded is "then draws a card. Put a +1/+1 counter on this creature".
// An on-disk key.
var memoryWormDiscarded = game.RegisterDiscardThen("choose-cards/memory-worm-discarded", func(g *game.Game, r game.DiscardsDone) error {
	if err := g.DrawNForEffect(r.Player, 1); err != nil {
		return err
	}
	if r.Item == nil || !effects.SourceIsStillThisPermanent(g, r.Item) {
		return nil
	}
	return effects.PlusOneCountersOnThis(1)(g, r.Item)
})
