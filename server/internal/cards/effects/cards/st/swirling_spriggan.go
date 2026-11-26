package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swirling Spriggan — {2}{G}{G} Creature — Goblin Shaman, 3/3:
//
//	"{G/U}{G/U}: Target creature you control becomes the color or colors of your
//	 choice until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b0cea17-9baf-43c7-9f01-44f255e9d687",
		Name:         "Swirling Spriggan",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G/U}{G/U}: Target creature you control becomes the color or colors of your choice until end of turn.",
			Cost:    effects.ManaCost("{G/U}{G/U}"),
			Targets: effects.TargetCreatureYouControl(),
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if len(ctx.LegalTargets()) == 0 {
					return nil
				}
				effects.ChooseColorsTail(game.ColorForBenefit, g, item.Controller, item.SourceCardID,
					"Swirling Spriggan — choose one or more colors", ctx.ItemTail(swirlingSpriggan523ColorsChosen))
				return nil
			},
		}},
	})
}

var swirlingSpriggan523ColorsChosen = effects.CardTail("swirling-spriggan/colors-chosen", func(ctx *effects.Context, o game.TailOutcome) error {
	colors := game.ColorSetLetters(o.Tail.Word)
	if len(colors) == 0 {
		return nil
	}
	return effects.Gap2B05TargetBecomesColor(ctx, colors, "Swirling Spriggan — "+game.ColorSetName(o.Tail.Word)+" until end of turn")
})
