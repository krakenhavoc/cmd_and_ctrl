package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Volcanic Rush — Instant {4}{R} (Dragons of Tarkir, #5551):
//
//	"Attacking creatures get +2/+0 and gain trample until end of turn."
//
// Trumpet Blast's attacking-creatures pump with trample added. The set of
// attackers is fixed as the spell resolves (CR 611.2c).
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "6944261c-9f35-4a51-9b1a-c29e24f292e1",
		Name:         "Volcanic Rush",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.BoostUntilEOT{Match: effects.AttackingCreature(), Power: 2, Label: "Volcanic Rush — attacking creatures get +2/+0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Match: effects.AttackingCreature(), Keywords: []string{"trample"}, Label: "Volcanic Rush — trample"}.Apply(ctx)
		},
	})
}
