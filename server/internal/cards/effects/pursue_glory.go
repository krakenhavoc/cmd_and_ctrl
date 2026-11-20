package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f85edb4f-1eec-40cc-9e21-0a6fb71e8499",
		Name:         "Pursue Glory",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BoostUntilEOT{Match: AttackingCreature(), Power: 2, Label: "Pursue Glory — attacking creatures get +2/+0"}.Apply(ctx)
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
