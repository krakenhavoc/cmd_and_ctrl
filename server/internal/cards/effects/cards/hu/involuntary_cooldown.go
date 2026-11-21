package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "4640ce0e-f82f-4bfd-93fc-2d865863fc2e",
		Name:         "Involuntary Cooldown",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("up to two target artifacts and/or creatures", effects.Or(effects.Artifact(), effects.Creature())).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, id := range effects.LegalTargetCardIDs(ctx) {
				if err := (effects.TapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.AddCounter{Target: id, Kind: game.CounterStun, N: 2}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
