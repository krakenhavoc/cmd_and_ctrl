package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wizard's Retort — Instant {1}{U}{U}:
//
//	"This spell costs {1} less to cast if you control a Wizard.
//	 Counter target spell."
//
// #746: a conditional self cost modifier.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b828251c-86a9-454f-9852-d0876d0f5153",
		Name:         "Wizard's Retort",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(1, "This spell costs {1} less to cast if you control a Wizard.", effects.YouControlA(effects.HasSubtype("Wizard"))),
		},
		OnResolve: effects.CounterTheTargetSpell,
	})
}
