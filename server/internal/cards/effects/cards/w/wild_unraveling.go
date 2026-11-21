package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wild Unraveling — Instant {U}{U}:
//
//	"As an additional cost to cast this spell, blight 2 or pay {1}. (To
//	 blight 2, put two -1/-1 counters on a creature you control.) Counter
//	 target spell."
//
// The either/or cost (ADR 0100 §2), with the #1703 blight component in a
// branch.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd1a4219-437b-494f-ae0e-1d3181ffcd63",
		Name:         "Wild Unraveling",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.EitherCost(
			effects.BlightCost(2).Keyed("blight"),
			effects.ManaAdditionalCost("{1}").Keyed("mana"),
		),
		Targets:   effects.TargetSpell("target spell"),
		OnResolve: effects.CounterTheTargetSpell,
	})
}
