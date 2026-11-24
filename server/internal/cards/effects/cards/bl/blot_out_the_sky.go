package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blot Out the Sky — Sorcery {X}{W}{B} (Strixhaven: School of Mages, #5502):
//
//	"Create X tapped 2/1 white and black Inkling creature tokens with flying.
//	 If X is 6 or more, destroy all noncreature, nonland permanents."
//
// Welcome the Dead's tapped tokens, counted by the announced X (CR 107.3).
// The wipe is a second sentence checked as the spell resolves, after the
// tokens exist: it destroys every noncreature, nonland permanent (artifacts,
// enchantments, planeswalkers and battles; an artifact creature is a creature
// and stays). The tokens are creatures, so the wipe never touches them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ded78090-c6a1-4c9d-b7b9-b433b1d963b6",
		Name:         "Blot Out the Sky",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose: game.Purpose{
			Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true},
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			if x > 0 {
				tmpl := effects.TokenCard("2/1 white and black Inkling with flying")
				tmpl.Tapped = true
				if err := (effects.CreateToken{Template: tmpl, N: x}).Apply(ctx); err != nil {
					return err
				}
			}
			if x >= 6 {
				return effects.DestroyAllMatching{Match: effects.And(effects.Noncreature(), effects.Nonland())}.Apply(ctx)
			}
			return nil
		},
	})
}
