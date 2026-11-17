package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Involuntary Cooldown — Sorcery {3}{U} (The Brothers' War):
//
//	"Tap up to two target artifacts and/or creatures. Put two stun counters
//	 on each of them."
//
// Each target is tapped and then gets its counters, one at a time (Collector's
// Case's shape); a target that became illegal is skipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4640ce0e-f82f-4bfd-93fc-2d865863fc2e",
		Name:         "Involuntary Cooldown",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("up to two target artifacts and/or creatures", Or(Artifact(), Creature())).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, id := range legalTargetCardIDs(ctx) {
				if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
				if err := (AddCounter{Target: id, Kind: game.CounterStun, N: 2}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
