package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Agent of Shauku — Creature — Human Mercenary {1}{B}, 1/1 (Prophecy, S82 OLD
// batch 674, #5617):
//
//	"{1}{B}, Sacrifice a land: Target creature gets +2/+0 until end of turn."
//
// The mana and the land are both the cost (Spitting Spider's
// B08SacrificeALand beside ManaCost); the target is any creature, chosen as
// the ability is activated and checked again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2e0d739-b441-4173-a8f1-b831a19ea98e",
		Name:         "Agent of Shauku",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice a land: Target creature gets +2/+0 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0)),
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.B08SacrificeALand()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: 2, Label: "Agent of Shauku — +2/+0"}.Apply(ctx)
			},
		}},
	})
}
