package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quickchange — Instant {1}{U} (Ravnica: City of Guilds):
//
//	"Target creature becomes the color or colors of your choice until end of turn.
//	 Draw a card."
//
// Swirling Spriggan's color choice (the "color or colors of your choice" prompt,
// #6146): the caster picks one or more colors as the spell resolves, and the
// target becomes exactly those until end of turn (CR 105.3, layer 5), replacing
// its own. The draw follows the choice (a keyed continuation, so the prompt is a
// restore point). A target gone by resolution makes the whole spell do nothing,
// the draw included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77f34281-6f93-4043-a13d-8d83014e7620",
		Name:         "Quickchange",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerProtect, Draws: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(ctx.LegalTargets()) == 0 {
				return nil
			}
			effects.ChooseColorsTail(game.ColorForBenefit, ctx.Game, item.Controller, item.SourceCardID,
				"Quickchange — choose one or more colors", ctx.ItemTail(quickchangeColorsChosen))
			return nil
		},
	})
}

// quickchangeColorsChosen applies the chosen colors to the target, then draws.
// An on-disk key: never renamed, never reused.
var quickchangeColorsChosen = effects.CardTail("quickchange/colors-chosen", func(ctx *effects.Context, o game.TailOutcome) error {
	if colors := game.ColorSetLetters(o.Tail.Word); len(colors) > 0 {
		if err := effects.Gap2B05TargetBecomesColor(ctx, colors, "Quickchange — "+game.ColorSetName(o.Tail.Word)+" until end of turn"); err != nil {
			return err
		}
	}
	return effects.DrawCards{N: 1}.Apply(ctx)
})
