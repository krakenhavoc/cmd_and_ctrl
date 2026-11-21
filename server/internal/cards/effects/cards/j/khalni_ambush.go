package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Khalni Ambush // Khalni Territory — Modal DFC. Front: Instant
// {2}{G}: "Target creature you control fights target creature you don't
// control." Back: Land (mdfc_lands.go).
//
// Two target clauses, one per slot. The fight is the shared one
// (b10Fight, CR 701.14): each deals damage equal to its power to the
// other as one damage instance, and nothing happens if either has left
// the battlefield or either target became illegal. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "37a55560-6e32-4f54-b9a8-fd157aea6eb5",
		Name:         "Khalni Ambush",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: effects.ResolveSteps(effects.Instruction(effects.FightTargetClauses)),
	})
}
