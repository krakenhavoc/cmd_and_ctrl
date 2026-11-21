package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crimson Wisps — Instant {R} (Foundations Jumpstart, #5929):
//
//	"Target creature becomes red and gains haste until end of turn.
//	 Draw a card."
//
// Cerulean Wisps' colour change (a layer 5 replacement of its colours,
// CR 105.3) in red, with haste granted for the turn, then the draw. A
// target that is gone as the spell resolves fizzles the whole spell,
// draw included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e699d140-eb13-4f15-aecb-85c76e8b1b14",
		Name:         "Crimson Wisps",
		Purpose:      crimsonWispsPurpose(),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := effects.UntilEndOfTurn(ctx, target, nil,
				"Crimson Wisps — that creature becomes red",
				game.SetColorsMod(ctx.TextColors("R")...)); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"haste"}, Label: "Crimson Wisps — haste"}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}

// crimsonWispsPurpose is the haste the target gets until end of turn and the
// card drawn.
func crimsonWispsPurpose() game.Purpose {
	p := effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste"))
	p.Draws = 1
	return p
}
