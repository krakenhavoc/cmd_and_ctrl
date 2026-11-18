package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Behold the Unspeakable // Vision of the Unspeakable — a transforming
// Saga (Kamigawa: Neon Dynasty, #5493).
//
//	I   — Creatures you don't control get -2/-0 until your next turn.
//	II  — If you have one or fewer cards in hand, draw four cards.
//	      Otherwise, scry 2, then draw two cards.
//	III — Exile this Saga, then return it to the battlefield transformed
//	      under your control.
//
// Vision of the Unspeakable, Enchantment Creature — Spirit 0/0: "Flying,
// trample. This creature gets +1/+1 for each card in your hand."
//
// Chapter I affects the creatures on the battlefield as it resolves
// (CR 611.2c); one that enters later is not shrunk. Chapter II reads
// your hand as it resolves; the draw after the scry waits for the scry's
// answer (CR 608.2c).
func init() {
	Register(Spec{
		OracleID:     beholdTheUnspeakableOracleID,
		Name:         "Behold the Unspeakable",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 2, Draws: 2},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel("Behold the Unspeakable", 1, "creatures you don't control get -2/-0 until your next turn"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return ScopedEffectFor{
						Match:    And(Creature(), OpponentControls()),
						Mods:     []game.Mod{game.ModifyPTMod(-2, 0)},
						Duration: DurationUntilYourNextTurn(ctx, item.Controller),
						Label:    "Behold the Unspeakable — creatures you don't control get -2/-0",
					}.Apply(ctx)
				}),
			ChapterTrigger(2, SagaChapterLabel("Behold the Unspeakable", 2, "draw four if you have one or fewer cards in hand, otherwise scry 2, then draw two"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					p := g.PlayerByIDForEffect(item.Controller)
					if p != nil && p.Hand != nil && p.Hand.Size() <= 1 {
						return DrawCards{N: 4}.Apply(ctx)
					}
					return Scry{N: 2, ThenTail: ctx.ItemTail(beholdTheUnspeakableDrawTwo)}.Apply(ctx)
				}),
			ChapterTrigger(3, SagaChapterLabel("Behold the Unspeakable", 3, "exile it, then return it transformed"),
				ChapterExileAndReturnTransformed),
		},
	})

	Register(Spec{
		OracleID:        beholdTheUnspeakableOracleID + "#1",
		Name:            "Vision of the Unspeakable",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Static: []game.StaticAbility{
			s82NeoB18SelfPT("This creature gets +1/+1 for each card in your hand.",
				func(g *game.Game, source *game.Card) int {
					if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Hand != nil {
						return p.Hand.Size()
					}
					return 0
				}),
		},
	})
}

// beholdTheUnspeakableOracleID is shared with the back face ("#1").
const beholdTheUnspeakableOracleID = "464372ef-0b16-4b8d-ba6f-fbf5c905c479"

// beholdTheUnspeakableDrawTwo is the "then draw two cards" after the scry.
var beholdTheUnspeakableDrawTwo = CardTail("tail/s82-neo-b18-behold-the-unspeakable-draw-two",
	func(ctx *Context, _ game.TailOutcome) error { return DrawCards{N: 2}.Apply(ctx) })
