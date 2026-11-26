package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Retaliator Griffin — Creature — Griffin {1}{R}{G}{W}, 2/2 (Alara Reborn,
// #5586):
//
//	"Flying
//	 Whenever a source an opponent controls deals damage to you, you may put
//	 that many +1/+1 counters on this creature."
//
// Michiko Konda's damage watcher (ASourceAnOpponentControlsDealtDamageToYou: the
// damage is to the player, from a source an opponent controls), with the amount
// of the damage taken from the event as the trigger is put on the stack. The
// counters are optional and go on the Griffin only while it is still the
// permanent that triggered.
//
// No simplification.
func init() {
	const label = "Retaliator Griffin — you may put that many +1/+1 counters on this creature"
	effects.Register(effects.Spec{
		OracleID:        "dd4b9e5a-b3ab-4b70-b200-228c8add9f52",
		Name:            "Retaliator Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(game.TriggeredAbility{
				Watches:   []game.EventKind{game.EventDealDamage},
				AppliesTo: effects.ASourceAnOpponentControlsDealtDamageToYou,
				Key:       label,
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, label)
					item.Params.Amount = ev.Amount
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					if item.Params.Amount <= 0 || !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: item.Params.Amount}.Apply(effects.NewContext(g, item))
				},
			}, "Retaliator Griffin — put +1/+1 counters on it?"),
		},
	})
}
