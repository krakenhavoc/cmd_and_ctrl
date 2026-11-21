package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reclaiming Vines — Sorcery {2}{G}{G} (Battle for Zendikar, #5547):
//
//	"Destroy target artifact, enchantment, or land."
//
// Creeping Mold's sentence: one target among artifacts, enchantments
// and lands, checked again at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ec3d99d-2c49-469a-81d5-f1609b6e49ca",
		Name:         "Reclaiming Vines",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact, enchantment, or land", effects.Or(effects.Artifact(), effects.Enchantment(), effects.Land())),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
