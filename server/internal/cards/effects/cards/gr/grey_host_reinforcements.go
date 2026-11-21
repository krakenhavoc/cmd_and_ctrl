package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grey Host Reinforcements — Creature — Spirit Soldier {3}{W}, 1/1 (Tales
// of Middle-earth Commander):
//
//	"Flying, ward {3}
//	 When this creature enters, exile target player's graveyard. Put a
//	 number of +1/+1 counters on this creature equal to the number of
//	 creature cards exiled this way."
//
// Sins of the Father's graveyard exile: the whole graveyard goes at once
// and the tail counts the creature cards among the cards that really
// reached exile, as they were in the graveyard (CR 608.2c). The counters
// go on the Reinforcements, and none if it has left the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "053dff47-25df-424b-8328-3ee00739b4fd",
		Name:            "Grey Host Reinforcements",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{3}"), "Grey Host Reinforcements — ward {3}"),
			effects.Targeting(effects.WhenThisEnters("Grey Host Reinforcements — exile target player's graveyard; put a +1/+1 counter on this creature for each creature card exiled this way",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					p := ctx.PlayerByID(t.ID)
					if p == nil || p.Graveyard == nil {
						return nil
					}
					ids := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
					for _, c := range p.Graveyard.Cards {
						ids = append(ids, c.InstanceID)
					}
					return effects.ExileCardsThen(ctx, ids, s82GapB118GreyHostCounters)
				}),
				effects.TargetPlayer("target player")),
		},
	})
}

var s82GapB118GreyHostCounters = effects.GraveyardExileTail("grey-host-reinforcements/counters-per-creature",
	func(ctx *effects.Context, exiled []uuid.UUID, wasCreature map[uuid.UUID]bool) error {
		n := effects.CreatureCardsAmong(exiled, wasCreature)
		if n == 0 || !effects.OnBattlefield(ctx.Game, ctx.Source()) {
			return nil
		}
		return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: n}.Apply(ctx)
	})
