package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goliath Truck — Artifact — Vehicle {4}, 4/4 (Warhammer 40,000 Commander,
// S82 GAP batch 186, #5948):
//
//	"Stowage — Whenever this Vehicle attacks, put two +1/+1 counters on
//	 another target attacking creature.
//	 Crew 2"
//
// The target is chosen as the trigger goes on the stack (CR 603.3d) and
// must be an attacking creature other than the Truck.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d94144b0-8967-4904-bb6b-44b0022d1f19",
		Name:         "Goliath Truck",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Goliath Truck"),
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Goliath Truck — put two +1/+1 counters on another target attacking creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok {
						return nil
					}
					return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
				}),
				effects.Another(effects.TargetCreature("another target attacking creature", effects.IsAttackingCreature()))),
		},
	})
}
