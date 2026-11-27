package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Headlong Rush — Instant {1}{R} (Urza's Saga, S82 OLD batch 729, #5624):
//
//	"Attacking creatures gain first strike until end of turn."
//
// Every creature attacking as the spell resolves, whoever controls it, gains
// first strike; the set is fixed then (CR 611.2c), so a creature that attacks
// later in the turn gets none. Overrun's grant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc3abacc-6643-49c5-8df8-5b56ab20fba5",
		Name:         "Headlong Rush",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.AttackingCreature()),
				Keywords: []string{"first strike"},
				Label:    "Headlong Rush — first strike",
			}.Apply(ctx)
		},
	})
}
