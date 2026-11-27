package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Distorting Lens — Artifact {2} (Eighth Edition, S82 OLD batch 612, #5607):
//
//	"{T}: Target permanent becomes the color of your choice until end of turn."
//
// Illusion's colour change (Gap2B05TargetBecomesColor) on a permanent: the colour
// is chosen as the ability resolves, then set in layer 5 until end of turn
// (CR 613.1e), replacing every colour the permanent had. A target that is gone
// by then does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c9579c5-7edf-41ca-8595-713ca9245576",
		Name:         "Distorting Lens",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target permanent becomes the color of your choice until end of turn.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target permanent"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if len(ctx.LegalTargets()) == 0 {
					return nil
				}
				effects.ChooseColorTail(game.ColorForBenefit, g, item.Controller, item.SourceCardID,
					"Distorting Lens — choose a color", ctx.ItemTail(distortingLensColorChosen))
				return nil
			},
		}},
	})
}

// distortingLensColorChosen sets the chosen colour. An on-disk key.
var distortingLensColorChosen = effects.CardTail("distorting-lens/color-chosen", func(ctx *effects.Context, o game.TailOutcome) error {
	color := o.Tail.Word
	if color == "" {
		return nil
	}
	return effects.Gap2B05TargetBecomesColor(ctx, []string{color}, "Distorting Lens — "+game.ColorName(color)+" until end of turn")
})
