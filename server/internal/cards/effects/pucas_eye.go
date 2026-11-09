package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Puca's Eye — Artifact {2} (Lorwyn Eclipsed, #4458):
//
//	"When this artifact enters, draw a card, then choose a color. This
//	 artifact becomes the chosen color.
//	 {3}, {T}: Draw a card. Activate only if there are five colors among
//	 permanents you control."
//
// The enters trigger draws, then asks for the colour as it resolves
// (ChooseColorTail), and the answer is Mondo Gecko's colour-setting mod
// (layer 5, CR 105.3: that colour and no other) on the Eye with no end
// (game.IndefiniteDuration), pinned to the Eye as it is now, so an Eye
// that has left the battlefield is not changed and a new Eye is
// colourless again. The draw comes first either way.
//
// The activation condition is Case of the Shattered Pact's
// fiveColorsAmongYourPermanents, read through effective colours, so the
// Eye's own chosen colour counts toward the five.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "499b4cf9-2aaa-45a0-8362-77b727b7f7cc",
		Name:         "Puca's Eye",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Puca's Eye — draw a card, then choose a color; this artifact becomes that color",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
						return err
					}
					ChooseColorTail(game.ColorForBenefit, g, item.Controller, item.SourceCardID,
						"Puca's Eye — choose a color", ctx.ItemTail(pucasEyeColorChosen))
					return nil
				}),
		},
		Activated: []ActivatedAbility{{
			Label:     "{3}, {T}: Draw a card. Activate only if there are five colors among permanents you control.",
			Purpose:   game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:      Plus(ManaCost("{3}"), TapCost()),
			Condition: fiveColorsAmongYourPermanents,
			Effect:    Do(DrawCards{N: 1}),
		}},
	})
}

// pucasEyeColorChosen makes the Eye the chosen colour from now on. Run
// against the enters trigger. An on-disk key.
var pucasEyeColorChosen = CardTail("then/pucas-eye-color-chosen", func(ctx *Context, o game.TailOutcome) error {
	color := o.Tail.Word
	if color == "" || !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return ScopedEffectFor{
		Target:   ctx.Item.SourceCardID,
		Mods:     []game.Mod{game.SetColorsMod(color)},
		Duration: game.IndefiniteDuration(),
		Label:    "Puca's Eye — becomes " + game.ColorName(color),
	}.Apply(ctx)
})
