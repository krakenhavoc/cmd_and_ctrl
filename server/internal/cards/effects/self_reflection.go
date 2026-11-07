package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Self-Reflection — Sorcery {4}{U}{U} (The Lost Caverns of Ixalan, #4470):
//
//	"Create a token that's a copy of target creature you control.
//	 Flashback {3}{U}"
//
// Cackling Counterpart at sorcery speed with a cheaper flashback: the
// token copies only the creature's copiable values (TokenCopyOfSingleTarget),
// and a target that is gone or no longer legal as the spell resolves
// gives nothing (CR 608.2b). Flashback casts it from the graveyard and
// exiles it afterwards (CR 702.34a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "5c68853e-4fb8-465e-b262-e6dbaf972533",
		Name:             "Self-Reflection",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{U}")},
		Targets:          TargetCreature("target creature you control", YouControl()),
		OnResolve:        TokenCopyOfSingleTarget,
	})
}
