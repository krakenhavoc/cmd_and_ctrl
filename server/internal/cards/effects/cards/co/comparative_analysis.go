package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Comparative Analysis — Instant {3}{U} (Oath of the Gatewatch, #5545):
//
//	"Surge {2}{U} (You may cast this spell for its surge cost if you or a
//	 teammate has cast another spell this turn.)
//	 Target player draws two cards."
//
// Surge is the shared gated alternative cost (CR 702.117a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "59b869b9-cb48-4152-ad91-52f582e2df5b",
		Name:             "Comparative Analysis",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Surge("{2}{U}")},
		Targets:          effects.TargetPlayer("target player"),
		Purpose:          effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 2}),
		OnResolve:        effects.TargetPlayerDrawsTwo,
	})
}
