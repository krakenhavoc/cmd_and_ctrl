package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b10_helpers.go — shared words for Kamigawa: Neon Dynasty batch 10
// (#5493), the "modified matters" cards. "Modified" is CR 700.9, read by the
// catalog's one predicate (b07IsModified).

// neoB10ModifiedCreatureCount is the number of modified creatures `player`
// controls, leaving out `except` (uuid.Nil leaves out none).
func neoB10ModifiedCreatureCount(g *game.Game, player, except uuid.UUID) int {
	return countControlled(g, player, func(c game.Card) bool {
		return c.InstanceID != except && c.IsCreature() && b07IsModified(g, c)
	})
}

// neoB10ModifiedCreature is "a modified creature" as a card predicate.
func neoB10ModifiedCreature(g *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.IsCreature() && b07IsModified(g, c)
}

// neoB10ModifiedCount is the cost-modifier count "for each modified
// creature you control".
func neoB10ModifiedCount(q game.CostQuery) int {
	return PermanentsYouControl(neoB10ModifiedCreature)(q)
}

// neoB10ModifiedKey is the cast-time value both "as you cast this spell"
// cards record (CR 601.2i): how many modified creatures the caster controlled.
const neoB10ModifiedKey = "modified_creatures"

// neoB10CastTimeModified records the number of modified creatures the caster
// controls as the spell is cast.
func neoB10CastTimeModified() []game.CastTimeValue {
	return []game.CastTimeValue{{
		Key: neoB10ModifiedKey,
		Value: func(g *game.Game, item *game.StackItem) int {
			return neoB10ModifiedCreatureCount(g, item.Controller, uuid.Nil)
		},
	}}
}

// neoB10CastModified reads the recorded count back at resolution. A game
// restored from before the value existed falls back to the board as it
// resolves.
func neoB10CastModified(ctx *Context) int {
	n, ok := ctx.CastValue(neoB10ModifiedKey)
	if !ok {
		n = neoB10ModifiedCreatureCount(ctx.Game, ctx.Controller(), uuid.Nil)
	}
	return max(n, 0)
}

// neoB10ControlsModified is "if you control a modified creature".
func neoB10ControlsModified(g *game.Game, player uuid.UUID) bool {
	return neoB10ModifiedCreatureCount(g, player, uuid.Nil) > 0
}
