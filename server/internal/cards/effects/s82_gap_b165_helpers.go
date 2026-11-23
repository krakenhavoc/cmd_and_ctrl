package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// NoTappedLands reports whether `you` controls no tapped lands (Martyr's
// Soul's "if you control no tapped lands"). A land creature is a land, a
// phased-out land is not on the battlefield. Caller must hold g.mu.
func NoTappedLands(g *game.Game, you uuid.UUID) bool {
	if you == uuid.Nil {
		return false
	}
	return len(PermanentsControlledByMatching(g, you, tappedLandPredicate)) == 0
}

func tappedLandPredicate(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Tapped && c.IsLand() }

// CondNoTappedLands is "if you control no tapped lands" as an intervening-if
// Condition, read for the item's controller as it resolves. It declares a
// read of the battlefield (game.FootprintCondBoard), as CondNoUntappedLands
// does.
type CondNoTappedLands struct{}

func (CondNoTappedLands) Holds(ctx *Context) bool {
	return NoTappedLands(ctx.Game, ctx.Controller())
}

func (CondNoTappedLands) Condition() (game.FootprintStep, bool) {
	return CondStep(game.FootprintCondBoard, 0), true
}
