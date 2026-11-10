package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fugitive Codebreaker — Creature — Goblin Rogue {1}{R} (Murders at Karlov
// Manor, #4469):
//
//	"Prowess, haste
//	 Disguise {5}{R}. This cost is reduced by {1} for each instant and
//	 sorcery card in your graveyard. (You may cast this card face down for
//	 {3} as a 2/2 creature with ward {2}. Turn it face up any time for its
//	 disguise cost.)
//	 When this creature is turned face up, discard your hand, then draw
//	 three cards."
//
// The reduction is the keyword's own (FaceUpCostLess, game.FaceUpCostReduction):
// it takes generic mana off {5}{R} as the cost is priced, never the {R}, and
// the {3} to cast it face down is not reduced. The trigger discards the
// whole hand, then draws three (a keyed "then", so the draw waits for any
// discard prompt); an empty hand still draws three (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "af2b36ee-2945-41d3-a186-8fd28f3cde16",
		Name:            "Fugitive Codebreaker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess", "haste"},
		AlternativeCosts: []game.AlternativeCost{
			FaceUpCostLess(Disguise("{5}{R}"), instantAndSorceryCardsInGraveyard),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisIsTurnedFaceUp("Fugitive Codebreaker — discard your hand, then draw three cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					p := g.PlayerByIDForEffect(item.Controller)
					if p == nil {
						return nil
					}
					if p.Hand.Size() == 0 {
						return DrawCards{N: 3}.Apply(ctx)
					}
					return g.DiscardRandomKeyedForEffect(item.Controller, p.Hand.Size(), ctx.ItemTail(fugitiveCodebreakerDraw))
				}),
		},
	})
}

var fugitiveCodebreakerDraw = CardTail("tail/fugitive-codebreaker-draw", func(ctx *Context, _ game.TailOutcome) error {
	return DrawCards{N: 3}.Apply(ctx)
})
