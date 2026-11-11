package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Notion Rain — Sorcery {1}{U}{B}:
//
//	"Surveil 2, then draw two cards. Notion Rain deals 2 damage to you."
//
// The draw and the damage are one keyed tail on the surveil, so neither
// happens before the player has answered the prompt. The damage is dealt
// by Notion Rain (the tail's source) to its controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f93dbe17-71d1-4542-96ff-81ac6da9513f",
		Name:         "Notion Rain",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Surveil{Player: ctx.Controller(), N: 2, ThenTail: ctx.Tail(notionRainThen)}.Apply(ctx)
		},
	})
}

// notionRainThen is "draw two cards. Notion Rain deals 2 damage to you."
// An on-disk key: never renamed, never reused.
var notionRainThen = CardTail("then/notion-rain-draw-and-damage", func(ctx *Context, _ game.TailOutcome) error {
	if err := ctx.Game.DrawNForEffect(ctx.Controller(), 2); err != nil {
		return err
	}
	return DealDamage{Source: ctx.Source(), Target: ctx.Controller(), Amount: 2}.Apply(ctx)
})
