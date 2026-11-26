package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trap Essence — Khans of Tarkir (S82 OLD b346, #5554):
//
//	Instant {G}{U}{R}: Counter target creature spell. Put two +1/+1 counters on up to
//	one target creature.
//
// Two target clauses: the spell is countered only while it is still a legal
// target, and the counters go on the second target on their own (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d64adb26-62c4-44aa-bc2e-d05bd1cf657b",
		Name:         "Trap Essence",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetSpell("target creature spell", effects.Creature()).Then(
			effects.UpToOneTargetCreature("up to one target creature")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) > 0 && ctx.IsTargetLegal(item.Targets[0]) {
				if err := effects.CounterTheTargetSpell(item, ctx); err != nil {
					return err
				}
			}
			if len(item.Targets) > 1 && ctx.IsTargetLegal(item.Targets[1]) {
				return effects.AddCounter{Target: item.Targets[1].ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
			}
			return nil
		},
	})
}
