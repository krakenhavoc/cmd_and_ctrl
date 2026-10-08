package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	begin := AtBeginningOfYourCombat("Ornery Tumblewagg — put a +1/+1 counter on target creature", b36CounterOnChosenAnimal)
	begin.Targets = TargetCreature("target creature")

	attack := AttacksWhileSaddled("Ornery Tumblewagg — double the +1/+1 counters on target creature", func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		c, found := g.LookupCardForEffect(id)
		if !found || c.Counters[game.CounterPlusOne] <= 0 {
			return nil
		}
		return AddCounter{Target: id, Kind: game.CounterPlusOne, N: c.Counters[game.CounterPlusOne]}.Apply(ctx)
	})
	attack.Targets = TargetCreature("target creature")

	Register(Spec{
		OracleID:     "f8d7bcbd-9a1d-4fc6-abdd-88a7e01b9411",
		Name:         "Ornery Tumblewagg",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Saddle(2)},
		Triggered:    []game.TriggeredAbility{begin, attack},
	})
}
