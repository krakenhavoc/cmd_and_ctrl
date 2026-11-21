package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vitality Hunter — Creature — Nightmare {3}{W}, 3/4 (Commander 2020,
// #5513):
//
//	"Lifelink
//	 {X}{W}{W}: Monstrosity X. (If this creature isn't monstrous, put X +1/+1
//	 counters on it and it becomes monstrous.)
//	 When this creature becomes monstrous, put a lifelink counter on each of
//	 up to X target creatures."
//
// Polukranos, World Eater's shape: MonstrosityX off the announced X, and a
// becomes-monstrous trigger whose target clause is sized by that X (the N of
// the instruction, CR 701.37c, not the counters that landed). Each target
// still legal as the trigger resolves gets one lifelink counter (CR 122.1b);
// an X of 0 leaves the trigger with no targets to choose.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d383c589-d8c7-4c4d-8084-37034d14b385",
		Name:            "Vitality Hunter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		XMatters:        true,
		Activated:       []effects.ActivatedAbility{effects.MonstrosityX(effects.ManaCost("{X}{W}{W}"))},
		Triggered:       []game.TriggeredAbility{vitalityHunterTrigger()},
	})
}

func vitalityHunterTrigger() game.TriggeredAbility {
	t := effects.WhenBecomesMonstrous("Vitality Hunter — put a lifelink counter on each of up to X target creatures", vitalityHunterCounters)
	t.TargetsFrom = func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
		x := tc.Amount()
		if x <= 0 {
			return nil
		}
		return effects.TargetCreature("up to X target creatures").WithCount(0, x)
	}
	return t
}

// vitalityHunterCounters puts one lifelink counter on each legal target.
func vitalityHunterCounters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterLifelink, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
