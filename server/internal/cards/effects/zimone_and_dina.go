package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zimone and Dina — Legendary Creature — Human Dryad {B}{G}{U}, 3/4:
//
//	"Whenever you draw your second card each turn, target opponent loses 2
//	 life and you gain 2 life.
//	 {T}, Sacrifice another creature: Draw a card. You may put a land card
//	 from your hand onto the battlefield tapped. If you control eight or more
//	 lands, repeat this process once."
//
// The trigger is Alandra's "second card each turn" read off the turn's draw
// tally, with the drain Vein Ripper uses.
//
// The activated ability is one pass of "draw, then may put a land tapped",
// and the repeat is decided AFTER the first pass, as printed: the put is a
// prompt, so the check rides its keyed tail (the land is on the battlefield
// by then and counts toward the eight) and runs the pass once more with no
// further repeat. "Another creature" is object identity, so a second card
// with the same name can pay.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5bc66b18-22ac-4138-b527-fa711116e298",
		Name:         "Zimone and Dina",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b41DrewYourSecondCardThisTurn(ev, source, g)
				}, "Zimone and Dina — target opponent loses 2 life and you gain 2 life",
					func(g *game.Game, item *game.StackItem) error {
						return targetOpponentLosesAndYouGain(g, item, 2)
					}),
				TargetPlayer("target opponent", Opponent())),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice another creature: Draw a card. You may put a land card from your hand onto the battlefield tapped. If you control eight or more lands, repeat this process once.",
			Purpose: game.Purpose{Draws: 1, Lands: 1, Answers: game.AnswerSacOutlet},
			Cost:    Plus(TapCost(), SacrificeAnotherN(1, "another creature", Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return zimoneProcess(NewContext(g, item))
			},
		}},
	})
}

// zimonePass is one pass: draw a card, then may put a land card from hand
// onto the battlefield tapped. then is the put's keyed tail, if any.
func zimonePass(ctx *Context, then game.TailRef) error {
	if err := (DrawCards{Player: ctx.Controller(), N: 1}).Apply(ctx); err != nil {
		return err
	}
	put := MayPutALandFromHandTapped("Zimone and Dina")
	if then.Key() != "" {
		put.ThenTail = then
		put.ThenItem = true
	}
	return put.Apply(ctx)
}

// zimoneProcess is the whole ability: the first pass, whose tail decides the
// repeat.
func zimoneProcess(ctx *Context) error { return zimonePass(ctx, zimoneRepeat) }

// zimoneRepeat is "if you control eight or more lands, repeat this process
// once": checked after the first pass, and the second pass has no tail, so
// it is not repeated again.
var zimoneRepeat = PutFromHandTail("tail/s82-mom-b21-zimone-repeat", func(ctx *Context, res PutFromHandResult) error {
	lands := countControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsLand() })
	if lands < 8 {
		return nil
	}
	return zimonePass(ctx, game.TailRef{})
})
