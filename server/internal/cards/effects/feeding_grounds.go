package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feeding Grounds — Plane — Muraganda (#4870):
//
//	"Red spells cost {1} less to cast.
//	 Green spells cost {1} less to cast.
//	 Whenever chaos ensues, put X +1/+1 counters on target creature, where X
//	 is that creature's mana value."
//
// Two cost reductions over every player's spells (a red-green spell gets both,
// CR 601.2f), each of generic mana only. X is the target's mana value as the
// ability resolves (CR 202.3: a token with no mana cost is 0).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b1eff630-b56c-48a5-ac21-836a60157b35",
		Name:         "Feeding Grounds",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Red spells cost {1} less to cast.", ColoredSpell("R")),
			CostsLess(1, "Green spells cost {1} less to cast.", ColoredSpell("G")),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverChaosEnsues("Feeding Grounds — put X +1/+1 counters on target creature, where X is that creature's mana value",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					c, found := g.LookupCardForEffect(id)
					if !found {
						return nil
					}
					mv, _ := g.ManaValueForEffect(c)
					return AddCounter{Target: id, Kind: game.CounterPlusOne, N: mv}.Apply(ctx)
				}),
				TargetCreature("target creature")),
		},
	})
}
