package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// midB12DrawThenTransformThis is "Draw a card, then transform this …" (Ecstatic
// Awakener) and "Draw a card. Transform this artifact." (Mysterious Tome): the
// draw first, then the source turns over in place unless it already has
// since the ability went on the stack (ADR 0152 §3).
func MidB12DrawThenTransformThis(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
		return err
	}
	return TransformThis{}.Apply(ctx)
}
