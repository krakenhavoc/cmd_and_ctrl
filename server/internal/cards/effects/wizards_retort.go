package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wizard's Retort — Instant {1}{U}{U}:
//
//	"This spell costs {1} less to cast if you control a Wizard.
//	 Counter target spell."
//
// #746: a conditional self cost modifier.
func init() {
	Register(Spec{
		OracleID:     "b828251c-86a9-454f-9852-d0876d0f5153",
		Name:         "Wizard's Retort",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		SelfCostModifiers: []game.CostModifier{
			CostsLess(1, "This spell costs {1} less to cast if you control a Wizard.", YouControlA(HasSubtype("Wizard"))),
		},
		OnResolve: counterTheTargetSpell,
	})
}
