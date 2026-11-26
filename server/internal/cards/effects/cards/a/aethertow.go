package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aethertow — Instant {3}{W/U} (Shadowmoor, S82 older sets):
//
//	"Put target attacking or blocking creature on top of its owner's library.
//	 Conspire (As you cast this spell, you may tap two untapped creatures you
//	 control that share a color with it. When you do, copy it and you may choose a new target for the copy.)"
//
// Run Aground's tuck over an attacking or blocking creature.
// Conspire is the engine's optional cost and cast trigger (conspire.go, CR
// 702.78), so the card declares the cost only. The conspire copy is a copy of the spell and may choose a new target (CR 707.10c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "dd0cac88-0a78-408f-8eb0-080f79150180",
		Name:          "Aethertow",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		OptionalCosts: []game.AdditionalCost{effects.Conspire()},
		OnResolve:     effects.PutTargetOnTopOfOwnersLibrary,
	})
}
