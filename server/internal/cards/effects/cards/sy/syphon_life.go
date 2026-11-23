package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Syphon Life — Sorcery {1}{B}{B}:
//
//	"Target player loses 2 life and you gain 2 life.
//	 Retrace"
//
// Retrace (CR 702.81) is a graveyard cast for the PRINTED mana cost plus
// a land card discarded from hand: `CastableZones` opens the graveyard and
// `Retrace(cost)` prices it (ADR 0066, 2026-10-07 amendment). The spell goes
// back to the graveyard when it resolves, so it can be retraced again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "c9367dd3-1a50-4eae-993e-c055dc6c4dc5",
		Name:             "Syphon Life",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Retrace("{1}{B}{B}")},
		Targets:          effects.TargetPlayer("target player"),
		OnResolve:        effects.TargetPlayerLosesLifeYouGain(2),
	})
}
