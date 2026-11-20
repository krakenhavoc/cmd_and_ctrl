package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Saving Grasp — Instant {U} (Dark Ascension, #5571):
//
//	"Return target creature you own to your hand.
//	 Flashback {W} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// The target is a creature its caster OWNS, whoever controls it now
// (YouOwn). Flashback is the shared alternative cost (CR 702.34a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "f7a4458a-7e8a-46bb-b62d-991c606c7a15",
		Name:             "Saving Grasp",
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature you own", YouOwn()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{W}")},
		OnResolve:        bounceTheTarget,
	})
}
