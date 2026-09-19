package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Behold the Multiverse — {3}{U} Instant: "Scry 2, then draw two
// cards. Foretell {1}{U}"
//
// Preordain's shape with a bigger draw and foretell on top (#658,
// CR 702.143): {2} on a turn with spare mana, then {1}{U} at the end
// of somebody else's, which is what makes a four-mana instant a
// staple.
//
// The draw goes in Scry's `Then` rather than on the line after it,
// for the reason Preordain's comment gives: Apply only QUEUES the
// prompt, so a draw written afterwards happens before the player has
// put anything back — a different card, off a prompt that is then
// unanswerable because the drawn card is no longer in the library.
func init() {
	Register(Spec{
		OracleID:     "d7d2f701-77df-4169-bf98-0d51d6886e9b",
		Name:         "Behold the Multiverse",
		Completeness: CompletenessFull,
		SpecialActions: []game.SpecialAction{
			Foretell("{1}{U}"),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			return Scry{
				Player: controller,
				N:      2,
				Then: func(g *game.Game) error {
					return g.DrawNForEffect(controller, 2)
				},
			}.Apply(ctx)
		},
	})
}
