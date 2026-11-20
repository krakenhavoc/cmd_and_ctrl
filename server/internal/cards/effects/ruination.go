package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ruination — Sorcery {3}{R} (Mystery Booster 2, #5948):
//
//	"Destroy all nonbasic lands."
//
// Armageddon's sweep over the lands without the basic supertype (CR 205.4c):
// a Sacred Foundry is a Mountain and still nonbasic, a Snow-Covered Swamp is
// basic. One destruction event for every player's lands (CR 701.8); an
// indestructible land stays.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "59dbc873-72d9-4e7b-9442-2c7b8e8523e1",
		Name:         "Ruination",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: NonbasicLand()}.Apply(ctx)
		},
	})
}
