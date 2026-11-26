package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hydrolash — Instant {2}{U} (Game Night):
//
//	"Attacking creatures get -2/-0 until end of turn.
//	 Draw a card."
//
// Morale's sweep with a negative number: every attacking creature, whoever
// controls it, as the spell resolves (CR 611.2c), then the draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "31279757-499a-4d52-aa81-76b6f1983b53",
		Name:         "Hydrolash",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.BoostUntilEOT{Match: effects.AttackingCreature(), Power: -2, Label: "Hydrolash — attacking creatures get -2/-0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
