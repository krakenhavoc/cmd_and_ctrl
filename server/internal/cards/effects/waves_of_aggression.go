package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Waves of Aggression — Sorcery {3}{R/W}{R/W}:
//
//	"Untap all creatures that attacked this turn. After this main phase, there is an
//	 additional combat phase followed by an additional main phase.
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
		OracleID:         "10991b1a-7dd3-4fbf-a4c8-200dc62fc605",
		Name:             "Waves of Aggression",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Retrace("{3}{R/W}{R/W}")},
		OnResolve:        untapAttackersThenCombatAndMain,
	})
}
