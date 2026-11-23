package cardsp

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Parish-Blade Trainee — Creature — Human Soldier {1}{W}, 1/2 (Crimson Vow):
//
//	"Training (Whenever this creature attacks with another creature with
//	 greater power, put a +1/+1 counter on this creature.)
//	 When this creature dies, put its counters on target creature you
//	 control."
//
// Training is the shared Training() trigger. The death trigger targets as it
// goes on the stack and puts every counter the Trainee had as it last existed
// (CR 603.10a, 608.2h) on the target, kind by kind. With no other creature to
// target it is removed from the stack. The Trainee itself is in the graveyard,
// so it is never a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1733392-1402-458b-bbc7-673d733c731c",
		Name:         "Parish-Blade Trainee",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Training("Parish-Blade Trainee — training: put a +1/+1 counter on it"),
			effects.Targeting(effects.WhenThisDies("Parish-Blade Trainee — put its counters on target creature you control", parishBladeTraineeMoveCounters),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}

// parishBladeTraineeMoveCounters puts the dead Trainee's last-known counters
// on the chosen creature.
func parishBladeTraineeMoveCounters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target := effects.FirstLegalCardTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	left, ok := ctx.TriggeringPermanent()
	if !ok {
		return nil
	}
	counters := left.Counters
	kinds := make([]string, 0, len(counters))
	for kind, n := range counters {
		if n > 0 {
			kinds = append(kinds, kind)
		}
	}
	slices.Sort(kinds)
	for _, kind := range kinds {
		if err := g.AddCounterForEffect(target, kind, counters[kind]); err != nil {
			return err
		}
	}
	return nil
}
