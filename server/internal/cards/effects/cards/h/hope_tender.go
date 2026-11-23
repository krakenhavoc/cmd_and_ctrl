package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hope Tender — Creature — Human Druid {1}{G}, 2/2:
//
//	"{1}, {T}: Untap target land.
//	 {1}, {T}, Exert this creature: Untap two target lands. (An exerted
//	 creature won't untap during your next untap step.)"
//
// Both untap any land, yours or not. The second needs two different
// lands to target and pays an exert as well (ADR 0130 §4, CR 701.43a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f377da9-7d7b-4407-8846-75a12b4dd72d",
		Name:         "Hope Tender",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, {T}: Untap target land.",
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Targets: effects.TargetPermanent("target land", effects.Land()),
				Effect:  effects.B22UntapEachLegalTarget,
			},
			{
				Label:   "{1}, {T}, Exert this creature: Untap two target lands.",
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.ExertThis()),
				Targets: effects.TargetPermanent("two target lands", effects.Land()).WithCount(2, 2),
				Effect:  effects.B22UntapEachLegalTarget,
			},
		},
	})
}
