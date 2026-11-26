package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blood Tyrant — Creature — Vampire {4}{U}{B}{R}, 5/5 (Commander 2016, S82
// older sets, #5537):
//
//	"Flying, trample
//	 At the beginning of your upkeep, each player loses 1 life. Put a +1/+1
//	 counter on this creature for each 1 life lost this way.
//	 Whenever a player loses the game, put five +1/+1 counters on this
//	 creature."
//
// The counters are the life that was really lost (CR 119.3), told to a keyed
// tail once every loss, including any a replacement changed or stopped, has
// settled; the Tyrant must still be the same object when it does. A player
// leaving the game, for any reason, is "loses the game" (CR 104.3), so the
// five counters follow each elimination.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ac74bd52-a450-486f-8161-816fd8d966d0",
		Name:            "Blood Tyrant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Blood Tyrant — each player loses 1 life; put a +1/+1 counter on this creature for each 1 life lost this way",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ref, ok := ctx.SourceRef()
					if !ok {
						return nil
					}
					return g.LoseLifeEachKeyedForEffect(item.SourceCardID, effects.TablePlayers(ctx), 1, game.KeyedTail{
						Ref: bloodTyrantCounters, Player: item.Controller, Source: item.SourceCardID,
						Carry: []uuid.UUID{ref.ID}, N: ref.Epoch,
					})
				}),
			effects.On(game.EventPlayerEliminated, func(_ game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return true
			}, "Blood Tyrant — put five +1/+1 counters on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 5})),
		},
	})
}

// bloodTyrantCounters puts a +1/+1 counter on the Tyrant for each life lost.
// An on-disk key: never renamed.
var bloodTyrantCounters = game.RegisterTail("blood-tyrant/counters-for-life-lost", func(g *game.Game, o game.TailOutcome) error {
	if o.Amount <= 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	id := o.Tail.Carry[0]
	if !effects.OnBattlefield(g, id) || g.ObjectEpochForEffect(id) != o.Tail.N {
		return nil
	}
	return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: o.Amount}.Apply(
		effects.NewContext(g, &game.StackItem{Controller: o.Tail.Player, SourceCardID: o.Tail.Source}))
})
