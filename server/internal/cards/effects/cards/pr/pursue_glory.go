package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pursue Glory — Instant {3}{R}
// (Amonkhet, #5535):
//
//	"Attacking creatures get +2/+0 until end of turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Trumpet Blast, plus cycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f85edb4f-1eec-40cc-9e21-0a6fb71e8499",
		Name:         "Pursue Glory",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{Match: effects.AttackingCreature(), Power: 2, Label: "Pursue Glory — attacking creatures get +2/+0"}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
