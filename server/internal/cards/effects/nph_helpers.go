package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// nph_helpers.go — shared helpers for New Phyrexia (S82 #5575).

// nphThenItsControllerLosesOne is "<do something to target permanent>.
// Its controller loses 1 life." (Vapor Snag, Glissa's Scorn, Victorious
// Destruction). The permanent's controller is read before `act`, so the
// loss lands on the player who controlled it even when the permanent has
// left, or survived a destroy by indestructible or regeneration. A target
// gone by resolution does nothing (CR 608.2b).
func NphThenItsControllerLosesOne(act func(ctx *Context, id uuid.UUID) error) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		for _, t := range ctx.LegalTargets() {
			c, ok := ctx.Game.LookupCardForEffect(t.ID)
			if !ok {
				continue
			}
			controller := c.Controller
			if err := act(ctx, t.ID); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), controller, -1)
		}
		return nil
	}
}
