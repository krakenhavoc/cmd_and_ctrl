package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thermokarst — Sorcery {1}{G}{G} (Ice Age, S82 OLD, #5636):
//
//	"Destroy target land. If that land was a snow land, you gain 1 life."
//
// Icequake's shape with the rider gaining the caster 1 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f3b071a1-7501-4424-979a-780fa90eecb8",
		Name:         "Thermokarst",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve:    effects.S82OldB842Thermokarst,
	})
}
