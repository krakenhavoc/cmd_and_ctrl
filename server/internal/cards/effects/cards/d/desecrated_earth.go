package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desecrated Earth — Sorcery {4}{B} (Zendikar, S82 OLD b495):
//
//	"Destroy target land. Its controller discards a card."
//
// Old495DesecratedEarth: the land's controller is read as the spell resolves,
// before it is destroyed, and discards a card of their choice whether or not the
// land was destroyed (an indestructible one stays); a target that is no longer
// legal makes the whole spell do nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c9242fd-915b-4f6d-b4ee-830c9d80c9e8",
		Name:         "Desecrated Earth",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve:    effects.Old495DesecratedEarth,
	})
}
