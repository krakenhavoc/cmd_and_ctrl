package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunlance — Sorcery {W} (Planar Chaos, #5595):
//
//	"Sunlance deals 3 damage to target nonwhite creature."
//
// Lightning Bolt's body on a nonwhite creature: the colour is read at
// announcement and again as it resolves (CR 608.2b), so a creature that
// turned white in response is not dealt damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c8cc7e2-7bff-46e0-bb5b-1e721a56a3ef",
		Name:         "Sunlance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target nonwhite creature", effects.Not(effects.OfColor("W"))),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve:    effects.DamageToFirstTargetStep(3),
	})
}
