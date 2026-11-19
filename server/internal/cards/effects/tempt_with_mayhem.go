package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tempt with Mayhem — Instant {1}{R}{R} (Modern Horizons 3 Commander,
// #5948):
//
//	"Tempting offer — Choose target instant or sorcery spell. Each opponent
//	 may copy that spell and may choose new targets for the copy they
//	 control. You copy that spell once plus an additional time for each
//	 opponent who copied the spell this way. You may choose new targets for
//	 the copies you control."
//
// The tempting-offer helper (Tempt with Vengeance's): the caster's own copy
// first, then each opponent in turn order is asked, and one who accepts
// copies the spell and gives the caster one more. Every copy is controlled by
// the player who made it (CR 707.10) and that player may choose new targets
// for it (CR 707.10c). A target spell that has left the stack makes the whole
// spell fizzle (CR 608.2b), so nothing is copied.
//
// No simplification.
func init() {
	tempt := newCmdr2S49Tempting("confirm/tempt-with-mayhem",
		"Tempt with Mayhem — copy the spell (the caster copies it one more time if you do)?", temptWithMayhemCopy)
	Register(Spec{
		OracleID:     "cbf64c58-664f-431f-987e-08ae70d23f2b",
		Name:         "Tempt with Mayhem",
		Completeness: CompletenessFull,
		Targets:      instantOrSorcerySpell("target instant or sorcery spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return tempt.resolve(ctx)
		},
	})
}

// temptWithMayhemCopy is one copy of the target spell, made and controlled
// by `player`.
func temptWithMayhemCopy(ctx *Context, player uuid.UUID) error {
	if ctx.Item == nil || len(ctx.Item.Targets) == 0 {
		return nil
	}
	return CopySpell{
		StackID:          ctx.Item.Targets[0].ID,
		Controller:       player,
		ChooseNewTargets: true,
	}.Apply(ctx)
}
