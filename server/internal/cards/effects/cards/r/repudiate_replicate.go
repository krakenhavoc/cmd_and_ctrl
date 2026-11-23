package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Repudiate // Replicate — split card {G/U}{G/U} // {1}{G}{U} (Ravnica
// Remastered):
//
//	Repudiate (Instant): "Counter target activated or triggered ability.
//	 (Mana abilities can't be targeted.)"
//	Replicate (Sorcery): "Create a token that's a copy of target creature
//	 you control."
//
// Stifle's counter and Cackling Counterpart's token copy. The left half is
// the bare oracle ID and the right half "<oracle>#1" (ADR 0034 faces).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a43ecf8b-5dc9-4ac4-ae1f-b663aa8c9d7b",
		Name:         "Repudiate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.AbilityOnStack("target activated or triggered ability"),
		OnResolve:    effects.CounterTheTargetSpell,
	})
	effects.Register(effects.Spec{
		OracleID:     "a43ecf8b-5dc9-4ac4-ae1f-b663aa8c9d7b#1",
		Name:         "Replicate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve:    effects.TokenCopyOfSingleTarget,
	})
}
