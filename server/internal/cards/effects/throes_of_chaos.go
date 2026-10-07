package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Throes of Chaos — Sorcery {3}{R}:
//
//	"Cascade (When you cast this spell, exile cards from the top of your library
//	 until you exile a nonland card that costs less. You may cast it without paying
//	 its mana cost. Put the exiled cards on the bottom in a random order.)
//	 Retrace"
//
// Retrace (CR 702.81) is a graveyard cast for the PRINTED mana cost plus
// a land card discarded from hand: `CastableZones` opens the graveyard and
// `Retrace(cost)` prices it (ADR 0066, 2026-10-07 amendment). The spell goes
// back to the graveyard when it resolves, so it can be retraced again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "e3444fcf-70ed-4d6e-aea9-030af15cad56",
		Name:             "Throes of Chaos",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Retrace("{3}{R}")},
		Triggered:        []game.TriggeredAbility{Cascade()},
	})
}
