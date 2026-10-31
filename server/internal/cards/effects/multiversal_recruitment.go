package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Multiversal Recruitment — {3}{U} Sorcery (Marvel Super Heroes, #4455):
//
//	"Create a token that's a copy of target creature you control, except it
//	 isn't legendary.
//	 Flashback {5}{U}{U} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Quantum Misalignment's copy (nonLegendaryTokenCopyOfTheTarget) and the
// engine's flashback (CR 702.34).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "8c65d5bd-5d90-4e84-8251-cfef21dd6332",
		Name:             "Multiversal Recruitment",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{U}{U}")},
		Targets:          TargetCreature("target creature you control", YouControl()),
		OnResolve:        nonLegendaryTokenCopyOfTheTarget,
	})
}
