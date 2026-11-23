package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daring Demolition — Sorcery {2}{B}{B} (Aether Revolt,
// #5536):
//
//	"Destroy target creature or Vehicle."
//
// Spin Out at sorcery speed: an uncrewed Vehicle is a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c8a3962-42a5-4070-8719-00d7483c7dbe",
		Name:         "Daring Demolition",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or Vehicle", effects.Or(effects.Creature(), effects.HasSubtype("Vehicle"))),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
