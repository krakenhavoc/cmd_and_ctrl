package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oran-Rief Ooze — Creature — Ooze {2}{G}, 2/2 (Foundations Jumpstart):
//
//	"When this creature enters, put a +1/+1 counter on target creature
//	 you control.
//	 Whenever this creature attacks, put a +1/+1 counter on each
//	 attacking creature with a +1/+1 counter on it."
//
// The enter trigger targets a creature its controller controls, the Ooze
// included. The attack trigger reads the set as it resolves (CR 608.2h):
// every attacking creature that has a +1/+1 counter then, and each gets
// one counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94a7dcef-d1c1-4ffa-869e-b911e84ed66c",
		Name:         "Oran-Rief Ooze",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Oran-Rief Ooze — a +1/+1 counter on target creature you control",
				effects.PlusOneCounterOnFirstLegalTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
			effects.WheneverThisAttacks("Oran-Rief Ooze — a +1/+1 counter on each attacking creature with a +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					var ids []uuid.UUID
					for _, id := range effects.AttackingCreatures(g) {
						if c, ok := g.LookupCardForEffect(id); ok && c.Counters[game.CounterPlusOne] > 0 {
							ids = append(ids, id)
						}
					}
					return effects.B13PutCounterOnEach(effects.NewContext(g, item), ids)
				}),
		},
	})
}
