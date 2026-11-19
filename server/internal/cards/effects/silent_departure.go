package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silent Departure — Sorcery {U} (Foundations Jumpstart):
//
//	"Return target creature to its owner's hand.
//	 Flashback {4}{U} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// A bounce with the shared Flashback, as Abandon the Post pairs it with a
// target restriction.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "c41641f0-1abf-4776-9dec-1882f2d9badd",
		Name:             "Silent Departure",
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{4}{U}")},
		OnResolve:        bounceTheTarget,
	})
}
