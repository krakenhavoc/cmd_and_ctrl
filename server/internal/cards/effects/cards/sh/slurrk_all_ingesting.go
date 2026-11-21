package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slurrk, All-Ingesting — Legendary Creature — Ooze {5}{G}, 0/0:
//
//	"Slurrk enters with five +1/+1 counters on it.
//	 Whenever Slurrk or another creature you control dies, if it had a
//	 +1/+1 counter on it, put a +1/+1 counter on each creature you
//	 control that has a +1/+1 counter on it.
//	 Partner"
//
// The entry counters are the printed CR 614.1c clause. The intervening
// "if" (CR 603.4) reads the dead creature's last-known counters (CR
// 603.10a), which is a fact about that moment and cannot change. The
// set of creatures that get the counter is fixed as the ability
// resolves: those still on the battlefield under the controller's
// control that have a +1/+1 counter then. Partner is the deck-building
// rule the commander code reads from the oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a0892be5-3272-431e-bbed-fc6520b259a7",
		Name:         "Slurrk, All-Ingesting",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 5, "Slurrk: enters with five +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B16SelfOrAnotherCreatureYouControlDied(ev, source, g) &&
					g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne] > 0
			}, "Slurrk — put a +1/+1 counter on each creature you control that has a +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() && c.Counters[game.CounterPlusOne] > 0 {
							ids = append(ids, c.InstanceID)
						}
					}
					for _, id := range ids {
						if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
