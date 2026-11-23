package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rapid Decay — Instant {1}{B}:
//
//	"Exile up to three target cards from a single graveyard.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Scarab Feast with a different cycling cost (#1807, ADR 0106 §5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c769dfc-79be-4fcb-840c-b72cafdd0a23",
		Name:         "Rapid Decay",
		Completeness: effects.CompletenessFull,
		Targets:      effects.UpToNCardsFromASingleGraveyard(3),
		OnResolve:    effects.ExileTargetCardsOnResolve,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
