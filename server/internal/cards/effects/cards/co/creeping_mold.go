package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Creeping Mold — Sorcery {2}{G}{G} (Visions, #5631):
//
//	"Destroy target artifact, enchantment, or land."
//
// A target that has left, or is no longer any of the three, is illegal
// and the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "59180e94-ccdf-4d9f-9a4a-fe55497d0d63",
		Name:         "Creeping Mold",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact, enchantment, or land", effects.Or(effects.Artifact(), effects.Enchantment(), effects.Land())),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
