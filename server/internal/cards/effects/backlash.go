package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Backlash — Instant {1}{B}{R} (Invasion, #5616):
//
//	"Tap target untapped creature. That creature deals damage equal to
//	 its power to its controller."
//
// Ana Battlemage's kicked clause as a whole spell: the untappedCreature
// target clause and the shared
// tapTargetThenItDealsItsPowerToItsController body. The creature is the
// damage's source (so its lifelink, deathtouch or infect apply), its
// power is read after the tap, and "its controller" is whoever controls
// it as the spell resolves. A creature that became tapped in response is
// no longer a legal target (CR 608.2b), and the spell does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3db9bf7b-da4e-428f-aa5d-a0cd1afeb16c",
		Name:         "Backlash",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target untapped creature", untappedCreature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return tapTargetThenItDealsItsPowerToItsController(ctx.Game, item)
		},
	})
}
