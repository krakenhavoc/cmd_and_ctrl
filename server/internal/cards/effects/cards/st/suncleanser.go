package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Suncleanser — Creature — Human Cleric {1}{W}, 1/4 (Core Set 2019, S82 OLD batch 198, #5527):
//
//	"When this creature enters, choose one —
//	 • Remove all counters from target creature. It can't have counters put on
//	   it for as long as this creature remains on the battlefield.
//	 • Target opponent loses all counters. That player can't get counters for
//	   as long as this creature remains on the battlefield."
//
// One modal enters trigger, the mode chosen as it goes on the stack (CR
// 700.2a). The first mode strips every kind off the creature, then a layer-6
// restriction record on that one permanent (game.ObjectCantHaveCounters) holds
// "can't have counters" while Suncleanser stays; the second takes every kind
// off the opponent and writes #6241's player record
// (PlayerCantGetCountersWhileSourceRemains). Both durations never begin when
// Suncleanser has already left (#1432). With its target illegal a mode does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	trigger := effects.WhenThisEnters("Suncleanser — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = effects.ChooseOne(
		effects.ModeDoing("Remove all counters from target creature. It can't have counters put on it for as long as this creature remains on the battlefield.",
			effects.TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := effects.RemoveEveryCounterFrom(ctx, id); err != nil {
					return err
				}
				return effects.S82OldB198CantHaveCountersWhileSourceRemains(ctx, id, "Suncleanser — can't have counters put on it")
			}),
		effects.ModeDoing("Target opponent loses all counters. That player can't get counters for as long as this creature remains on the battlefield.",
			effects.TargetPlayer("target opponent", effects.Opponent()),
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				id, ok := effects.FirstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				if err := effects.S82OldB198PlayerLosesAllCounters(ctx, id); err != nil {
					return err
				}
				return effects.PlayerCantGetCountersWhileSourceRemains(ctx, id, "Suncleanser — can't get counters")
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "437661f7-a32f-45c9-a006-05a1e38ee8d9",
		Name:         "Suncleanser",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
