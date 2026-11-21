package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Selfless Police Captain — Creature — Human Detective {1}{W}, 1/1 (Marvel
// Super Heroes Commander, #4455):
//
//	"This creature enters with a +1/+1 counter on it.
//	 When this creature leaves the battlefield, put its +1/+1 counters on
//	 target creature you control."
//
// The entry counter is Afiya Grove's self-replacement (CR 614.1c). The
// leaves trigger reads the +1/+1 counters the Captain had as it left
// (CR 603.10a, the triggering permanent's last-known information, as The
// Ozolith's collector does), however it left, and puts them on the chosen
// creature as one placement.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ece77a64-8d99-430e-8407-bdb321828d63",
		Name:         "Selfless Police Captain",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 1, "Selfless Police Captain: enters with a +1/+1 counter"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisLeaves("Selfless Police Captain — put its +1/+1 counters on target creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					left, found := ctx.TriggeringPermanent()
					n := left.Counters[game.CounterPlusOne]
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !found || n <= 0 || !ok || !effects.OnBattlefield(g, id) {
						return nil
					}
					return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
				}),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
