package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Holy Day — Instant {W}:
//
//	"Prevent all combat damage that would be dealt this turn."
//
// Fog in white, word for word. Registered separately because the
// catalog keys on oracle ID and a white deck runs this one; the
// effect is the shared primitive.
func init() {
	Register(Spec{
		OracleID:     "98423a34-f044-4811-b288-56981d604b6e",
		Name:         "Holy Day",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return PreventAllCombatDamageThisTurn{
				Label: "Holy Day: prevent combat damage",
			}.Apply(ctx)
		},
	})
}
