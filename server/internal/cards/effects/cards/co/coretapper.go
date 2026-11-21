package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coretapper — Artifact Creature — Myr {2}, 1/1 (Foundations Commander):
//
//	"{T}: Put a charge counter on target artifact.
//	 Sacrifice this creature: Put two charge counters on target
//	 artifact."
//
// Two activated abilities with a target artifact each; the sacrifice is
// a cost, so the counters still land if the Myr is gone by resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66e37011-f3d2-41ed-8e09-7bba9464e8c3",
		Name:         "Coretapper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Put a charge counter on target artifact.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
				Effect:  cmdrB9ChargeCountersOnTarget(1),
			},
			{
				Label:   "Sacrifice this creature: Put two charge counters on target artifact.",
				Cost:    effects.SacrificeThis(),
				Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
				Effect:  cmdrB9ChargeCountersOnTarget(2),
			},
		},
	})
}

// cmdrB9ChargeCountersOnTarget puts n charge counters on the still-legal
// target.
func cmdrB9ChargeCountersOnTarget(n int) effects.Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return effects.AddCounter{Target: t.ID, Kind: game.CounterCharge, N: n}.Apply(ctx)
			}
		}
		return nil
	}
}
