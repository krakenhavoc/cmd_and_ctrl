package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desert Twister — Sorcery {4}{G}{G} (Arabian Nights, #5641):
//
//	"Destroy target permanent."
//
// Vindicate's text at green's price: any permanent, lands included.
// Indestructible is honoured by the destroy path.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f880348-6dc8-4cf8-9313-9893c41a70a3",
		Name:         "Desert Twister",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target permanent"),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
