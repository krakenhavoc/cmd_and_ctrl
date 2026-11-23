package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ornery Tumblewagg — Creature — Brushwagg Mount {2}{G}:
//
//	"At the beginning of combat on your turn, put a +1/+1 counter on
//	 target creature.
//	 Whenever this creature attacks while saddled, double the number of
//	 +1/+1 counters on target creature.
//	 Saddle 2"
//
// Both triggers target (any creature, yours or not). The doubling reads
// the count as the ability resolves and places that many more, through
// AddCounter, so Doubling Season and Hardened Scales apply to it as
// printed. A creature with none gains none.
//
// No simplification.
func init() {
	begin := effects.AtBeginningOfYourCombat("Ornery Tumblewagg — put a +1/+1 counter on target creature", effects.B36CounterOnChosenAnimal)
	begin.Targets = effects.TargetCreature("target creature")

	attack := effects.AttacksWhileSaddled("Ornery Tumblewagg — double the +1/+1 counters on target creature", func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		id, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		c, found := g.LookupCardForEffect(id)
		if !found || c.Counters[game.CounterPlusOne] <= 0 {
			return nil
		}
		return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: c.Counters[game.CounterPlusOne]}.Apply(ctx)
	})
	attack.Targets = effects.TargetCreature("target creature")

	effects.Register(effects.Spec{
		OracleID:     "f8d7bcbd-9a1d-4fc6-abdd-88a7e01b9411",
		Name:         "Ornery Tumblewagg",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Saddle(2)},
		Triggered:    []game.TriggeredAbility{begin, attack},
	})
}
