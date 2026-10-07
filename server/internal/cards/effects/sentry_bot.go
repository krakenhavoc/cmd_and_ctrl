package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sentry Bot — Artifact Creature — Robot {4}{W}, 2/5:
//
//	"Flash
//	 This spell costs {1} less to cast for each creature attacking you.
//	 When this creature enters, you get {E} for each creature attacking
//	 you.
//	 At the beginning of combat on your turn, you may pay {E}{E}{E}. If
//	 you do, put a +1/+1 counter on each creature you control."
//
// ADR 0129 §3 (#1995). "Attacking you" is a creature attacking its
// controller as a player, not a planeswalker or battle they control
// (AttackingYou). The cost reduction counts them as the spell is cast;
// the energy counts them as the enters trigger resolves (CR 608.2h).
// The combat payment is made as its trigger resolves (CR 118.12), and
// the counters go on each creature its controller controls then.
//
// No simplification.
func init() {
	attackingYou := And(Creature(), AttackingYou())
	Register(Spec{
		OracleID:        "d0c78ee6-babb-410d-9d3a-0c1395596e56",
		Name:            "Sentry Bot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(PermanentsOnBattlefield(attackingYou),
				"This spell costs {1} less to cast for each creature attacking you."),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sentry Bot — you get {E} for each creature attacking you",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, c := range g.Battlefield.Cards {
						if attackingYou(g, item.Controller, c) {
							n++
						}
					}
					return GetEnergy{N: n}.Apply(NewContext(g, item))
				}),
			AtBeginningOfYourCombat("Sentry Bot — you may pay {E}{E}{E}",
				mayPayEnergyThen("Sentry Bot", 3, "put a +1/+1 counter on each creature you control",
					func(g *game.Game, item *game.StackItem) error {
						var mine []uuid.UUID
						for _, c := range g.Battlefield.Cards {
							if c.Controller == item.Controller && c.IsCreature() {
								mine = append(mine, c.InstanceID)
							}
						}
						for _, id := range mine {
							if err := g.AddCounterByForEffect(item.Controller, id, game.CounterPlusOne, 1); err != nil {
								return err
							}
						}
						return nil
					})),
		},
	})
}
