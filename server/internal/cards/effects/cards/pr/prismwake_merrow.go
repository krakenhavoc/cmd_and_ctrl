package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prismwake Merrow — Creature — Merfolk Wizard {2}{U}, 2/1 (Shadowmoor, #5591):
//
//	"Flash
//	 When this creature enters, target permanent becomes the color or
//	 colors of your choice until end of turn."
//
// Scuttlemutt's colour pick (#6146) on an enters trigger: as it resolves
// the controller names one or more colours and the target's colours are
// replaced by exactly those in layer 5 until end of turn (CR 105.3,
// 613.1e). A target that has left gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7d5126a9-ab81-4de6-817d-3b4b84c7ca5f",
		Name:            "Prismwake Merrow",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Prismwake Merrow — target permanent becomes the color or colors of your choice until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if len(ctx.LegalTargets()) == 0 {
						return nil
					}
					effects.ChooseColorsTail(game.ColorForBenefit, g, item.Controller, item.SourceCardID,
						"Prismwake Merrow — choose one or more colors", ctx.ItemTail(prismwakeMerrowColorsChosen))
					return nil
				}), effects.TargetPermanent("target permanent")),
		},
	})
}

// prismwakeMerrowColorsChosen sets the target to the chosen colours until end
// of turn. An on-disk key: never renamed, never reused.
var prismwakeMerrowColorsChosen = effects.CardTail("prismwake-merrow/colors-chosen", func(ctx *effects.Context, o game.TailOutcome) error {
	colors := game.ColorSetLetters(o.Tail.Word)
	if len(colors) == 0 {
		return nil
	}
	return effects.Gap2B05TargetBecomesColor(ctx, colors, "Prismwake Merrow — "+game.ColorSetName(o.Tail.Word)+" until end of turn")
})
