package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Servant of the Scale — Creature — Human Soldier {G}, 0/0 (Foundations
// Jumpstart, S82 BNS):
//
//	"This creature enters with a +1/+1 counter on it.
//	 When this creature dies, put X +1/+1 counters on target creature you
//	 control, where X is the number of +1/+1 counters on this creature."
//
// The entry counter is a replacement (Arwen's b10EntersWithCounters), so
// the Servant is never a 0/0 on the battlefield. X is read from the Servant
// as it last existed on the battlefield (CR 603.10a, 608.2h), counting only
// +1/+1 counters; with none the trigger does nothing. The target is chosen
// as the trigger goes on the stack and rechecked as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d404420-e3de-47a8-a159-2f634a73f2c9",
		Name:         "Servant of the Scale",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 1, "Servant of the Scale: enters with a +1/+1 counter on it"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Servant of the Scale — put X +1/+1 counters on target creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					if !ok || info.Counters[game.CounterPlusOne] <= 0 {
						return nil
					}
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return AddCounter{Target: id, Kind: game.CounterPlusOne, N: info.Counters[game.CounterPlusOne]}.Apply(ctx)
				}), TargetCreature("target creature you control", YouControl())),
		},
	})
}
