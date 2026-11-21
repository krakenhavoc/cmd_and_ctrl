package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Consign // Oblivion — split card with aftermath, oracle
// 8fe7436b-2467-486e-8a2d-7cbf21a65da9:
//
//	Consign — Instant {1}{U}: "Return target nonland permanent to its
//	          owner's hand."
//	Oblivion — Sorcery {4}{B}, aftermath: "Target opponent discards two
//	          cards."
//
// Insult // Injury's registration: Consign under the bare oracle ID,
// Oblivion under "#1" (ADR 0103). Aftermath is the engine's (CR 702.127a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8fe7436b-2467-486e-8a2d-7cbf21a65da9",
		Name:         "Consign",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve:    effects.BounceTheTarget,
	})
	effects.Register(effects.Spec{
		OracleID:     "8fe7436b-2467-486e-8a2d-7cbf21a65da9#1",
		Name:         "Oblivion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TargetOpponentDiscards(2)(ctx.Game, item)
		},
	})
}
