package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soulstinger — Creature — Scorpion Demon {3}{B}, 4/5 (Amonkhet, S82 OLD batch 242, #5535):
//
//	"When this creature enters, put two -1/-1 counters on target creature you
//	 control.
//	 When this creature dies, you may put a -1/-1 counter on target creature
//	 for each -1/-1 counter on this creature."
//
// The enters target may be the Scorpion itself. The dies trigger is a "you
// may" with a target chosen as it goes on the stack; the count is the -1/-1
// counters the Scorpion had as it died (last-known information, CR 603.10a,
// 608.2h), read as the ability resolves.
//
// No simplification.
func init() {
	dies := effects.Targeting(effects.Optional(effects.WhenThisDies("Soulstinger — put a -1/-1 counter on target creature for each -1/-1 counter on this creature",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			info, _ := ctx.TriggeringPermanent()
			n := info.Counters[game.CounterMinusOne]
			target, ok := ctx.ClauseTarget(0)
			if n <= 0 || !ok || target.Kind != game.TargetCard {
				return nil
			}
			return effects.AddCounter{Target: target.ID, Kind: game.CounterMinusOne, N: n}.Apply(ctx)
		}), "Soulstinger — put a -1/-1 counter on target creature for each -1/-1 counter on this creature?"),
		effects.TargetCreature("target creature"))
	effects.Register(effects.Spec{
		OracleID:     "31505942-52c4-4485-9241-c9592e6839e4",
		Name:         "Soulstinger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB242WhenThisEntersMinusCountersOnYourCreature("Soulstinger", 2),
			dies,
		},
	})
}
