package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fires of Undeath — Instant {2}{R} (Dark Ascension, #5571):
//
//	"Fires of Undeath deals 2 damage to any target.
//	 Flashback {5}{B} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Firebolt at instant speed with a black flashback cost (CR 702.34a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "205bcab6-8b6b-4d09-9890-0b7391d59788",
		Name:             "Fires of Undeath",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetAny(),
		Purpose:          effects.ForTargets(effects.DamageToTarget(0, 2)),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{5}{B}")},
		OnResolve:        effects.DamageToFirstTargetStep(2),
	})
}
