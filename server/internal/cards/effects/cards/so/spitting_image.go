package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spitting Image — Sorcery {4}{G/U}{G/U}:
//
//	"Create a token that's a copy of target creature.
//	 Retrace"
//
// Retrace (CR 702.81) is a graveyard cast for the PRINTED mana cost plus
// a land card discarded from hand: `CastableZones` opens the graveyard and
// `Retrace(cost)` prices it (ADR 0066, 2026-10-07 amendment). The spell goes
// back to the graveyard when it resolves, so it can be retraced again.
//
// No simplification: the token copy runs the ordinary entry pipeline, so
// the copied creature's enters-the-battlefield effects fire (see
// CreateTokenCopy).
func init() {
	effects.Register(effects.Spec{
		OracleID:         "a30dee74-86e3-4888-980e-b22437fbbb66",
		Name:             "Spitting Image",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Retrace("{4}{G/U}{G/U}")},
		Targets:          effects.TargetCreature("target creature"),
		OnResolve:        effects.TokenCopyOfSingleTarget,
	})
}
