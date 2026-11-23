package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Draconic Lore — Instant {5}{U} (Tarkir: Dragonstorm Commander):
//
//	"This spell costs {2} less to cast if you control a Dragon.
//	 Draw three cards."
//
// Terrific Team-Up's self cost modifier over a board read, judged when
// the cost is determined (CR 601.2f): a Dragon sacrificed afterwards
// (for mana, say) does not put the cost back. "Dragon" is a word of the
// card's text, read through the predicate's own text handling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "440881ad-48cf-4b82-8bb5-f506d1a51755",
		Name:         "Draconic Lore",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "This spell costs {2} less to cast if you control a Dragon.",
				effects.YouControlA(effects.Subtype("Dragon"))),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
	})
}
