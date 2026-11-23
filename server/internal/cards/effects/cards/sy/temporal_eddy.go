package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Temporal Eddy — Sorcery {2}{U}{U} (Time Spiral, #5596):
//
//	"Put target creature or land on top of its owner's library."
//
// Submerge's tuck over a creature or a land (putTargetOnTopOfOwnersLibrary).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a35e772-2547-41a3-a2a2-125b48c23aa5",
		Name:         "Temporal Eddy",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or land", effects.Or(effects.Creature(), effects.Land())),
		OnResolve:    effects.PutTargetOnTopOfOwnersLibrary,
	})
}
