package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Borne Upon a Wind — Instant {1}{U}:
//
//	"You may cast spells this turn as though they had flash.
//	 Draw a card."
//
// The flash grant is Vedalken Orrery's own permission, but with a
// DURATION rather than derived from a permanent's static ability:
// GrantCastTiming stores the statement on the controller for the
// rest of the turn (game/cast_timing.go, ADR 0066's effect half), the
// same primitive Emergence Zone and Winding Canyons use.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce19962d-94f9-4b2b-b668-963c0acce308",
		Name:         "Borne Upon a Wind",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (GrantCastTiming{
				Timing: game.TimingFlash,
				Label:  "Borne Upon a Wind — you may cast spells as though they had flash",
			}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
