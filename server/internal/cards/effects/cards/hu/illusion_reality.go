package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Illusion // Reality — Instant {U} // Instant {2}{G} (Apocalypse split
// card, S82 GAP2 b05 #5948, #6088):
//
//	Illusion: "Target spell or permanent becomes the color of your choice
//	 until end of turn."
//	Reality: "Destroy target artifact."
//
// The halves are ADR 0034 faces (the right under "<oracle>#1"); either is
// cast (CR 709.3). Illusion's colour is chosen as it resolves, then set in
// layer 5 until end of turn (CR 613.1e): on a spell through the stack
// step (game.SetSpellColorsForEffect, which a permanent spell hands to the
// permanent it becomes, CR 400.7a), on a permanent through an ordinary
// setColors record (Gap2B05TargetBecomesColor). Reality destroys its
// artifact.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5af2689d-c45c-4d45-b6a9-f32e8410b336",
		Name:         "Illusion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpellOrPermanent("target spell or permanent", nil, nil),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if len(ctx.LegalTargets()) == 0 {
				return nil
			}
			effects.ChooseColorTail(game.ColorForBenefit, ctx.Game, ctx.Controller(), ctx.Source(),
				"Illusion — choose a color", ctx.ItemTail(illusionColorChosen))
			return nil
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "5af2689d-c45c-4d45-b6a9-f32e8410b336#1",
		Name:         "Reality",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyChosenPermanent(ctx.Game, item)
		},
	})
}

// illusionColorChosen sets the chosen colour. An on-disk key.
var illusionColorChosen = effects.CardTail("illusion/color-chosen", func(ctx *effects.Context, o game.TailOutcome) error {
	color := o.Tail.Word
	if color == "" {
		return nil
	}
	return effects.Gap2B05TargetBecomesColor(ctx, []string{color}, "Illusion — "+game.ColorName(color)+" until end of turn")
})
