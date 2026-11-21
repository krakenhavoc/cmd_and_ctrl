package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drag to the Bottom — Sorcery {2}{B}{B}:
//
//	"Domain — Each creature gets -X/-X until end of turn, where X is 1
//	 plus the number of basic land types among lands you control."
//
// X is read once as the spell resolves (DomainCount, #5666) and fixed for
// the turn; the -X/-X lands on every creature on the battlefield then
// (CR 611.2c), yours included, as The Meathook Massacre's does. Its
// Purpose is a creature -X/-X sweep, partial since a large creature
// survives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c4d7a200-7401-4c9a-a7cf-6c98fa8eb61a",
		Name:         "Drag to the Bottom",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := 1 + effects.DomainCount(ctx.Game, ctx.Controller())
			return effects.BoostUntilEOT{
				Match:     effects.Creature(),
				Power:     -x,
				Toughness: -x,
				Label:     "Drag to the Bottom — each creature gets -X/-X",
			}.Apply(ctx)
		},
	})
}
