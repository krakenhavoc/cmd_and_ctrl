package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mondo Gecko — {1}{U}{U} Legendary Creature — Lizard Mutant 2/3 (Teenage
// Mutant Ninja Turtles, #4457):
//
//	"{1}, Discard a card: Until end of turn, Mondo Gecko becomes the color
//	 of your choice and gains hexproof from that color.
//	 Whenever Mondo Gecko deals combat damage to a player, draw a card for
//	 each color among permanents you control."
//
// The colour is chosen as the ability resolves (Skrelv, Defector Mite's
// ChooseColorTail), and the answer is one ScopedEffectFor until end of
// turn on Mondo Gecko: Cerulean Wisps' colour-setting mod (layer 5,
// CR 105.3: it becomes that colour and no other) and Skrelv's
// HexproofFromColor (CR 702.11d). A Gecko that left the battlefield, or
// came back as a new object, before the answer is not changed.
//
// The draw counts Faeburrow Elder's "each color among permanents you
// control" as the trigger resolves, so a colour the Gecko itself was
// given this turn counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa3ae120-756d-4c52-91f9-235767539d67",
		Name:         "Mondo Gecko",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Discard a card: Until end of turn, Mondo Gecko becomes the color of your choice and gains hexproof from that color.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    Plus(ManaCost("{1}"), DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				ChooseColorTail(game.ColorForProtection, g, item.Controller, item.SourceCardID,
					"Mondo Gecko — choose a color", ctx.ItemTail(mondoGeckoColorChosen))
				return nil
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Mondo Gecko — draw a card for each color among permanents you control",
				func(g *game.Game, item *game.StackItem) error {
					n := len(colorsAmongPermanentsYouControl(g, item.Controller))
					if n == 0 {
						return nil
					}
					return DrawCards{N: n}.Apply(NewContext(g, item))
				}),
		},
	})
}

// mondoGeckoColorChosen makes the Gecko the chosen colour, with hexproof
// from it, until end of turn. Run against the asking ability. An on-disk
// key.
var mondoGeckoColorChosen = CardTail("then/mondo-gecko-color-chosen", func(ctx *Context, o game.TailOutcome) error {
	color := o.Tail.Word
	if color == "" || !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return ScopedEffectFor{
		Target: ctx.Item.SourceCardID,
		Mods: []game.Mod{
			game.SetColorsMod(color),
			game.AddKeywordsMod(game.HexproofFromColor(color)),
		},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Mondo Gecko — " + game.ColorName(color) + ", hexproof from " + game.ColorName(color),
	}.Apply(ctx)
})
