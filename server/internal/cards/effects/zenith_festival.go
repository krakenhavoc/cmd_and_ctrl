package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zenith Festival — Sorcery {X}{R}{R} (#3931):
//
//	"Exile the top X cards of your library. You may play them until the
//	 end of your next turn.
//	 Harmonize {X}{R}{R}"
//
// Reckless Impulse's window (until the end of your next turn) over X
// cards. Harmonize (CR 702.180) includes {X}, so X is announced (CR
// 107.3a) and the tapped creature's power comes off it (CR 601.2f): the
// graveyard cast exiles more cards for the same mana. X of 0 exiles
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "4dabba8c-0490-41ec-a03e-f7d0c9f0edb6",
		Name:             "Zenith Festival",
		Completeness:     CompletenessFull,
		XMatters:         true,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Harmonize("{X}{R}{R}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			controller := item.Controller
			_, err := ctx.Game.ExileTopWithPermissionForEffect(controller, controller, x, game.CastPermission{
				Duration: ctx.Game.UntilEndOfYourNextTurnDuration(controller),
			})
			return err
		},
	})
}
