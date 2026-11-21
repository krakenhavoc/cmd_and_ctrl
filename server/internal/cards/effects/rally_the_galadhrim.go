package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rally the Galadhrim — Sorcery {2}{G}{U} (Tales of Middle-earth Commander):
//
//	"Create a token that's a copy of target creature you control.
//	 Conspire (As you cast this spell, you may tap two untapped
//	 creatures you control that share a color with it. When you do, copy
//	 it and you may choose a new target for the copy.)"
//
// The body is Cackling Counterpart's (TokenCopyOfSingleTarget, CR 707.2:
// copiable values only); conspire is the engine's optional cost and cast
// trigger (conspire.go, CR 702.78), so the card declares the cost only.
// The conspire copy is a copy of the spell and may choose a new target
// (CR 707.10c), so a conspired cast makes two token copies. A target
// that is gone as the spell resolves makes no token (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "84b2d0f6-ad98-4a71-a3d5-645b0f2db6c6",
		Name:          "Rally the Galadhrim",
		Completeness:  CompletenessFull,
		Targets:       TargetCreature("target creature you control", YouControl()),
		OptionalCosts: []game.AdditionalCost{Conspire()},
		OnResolve:     TokenCopyOfSingleTarget,
	})
}
