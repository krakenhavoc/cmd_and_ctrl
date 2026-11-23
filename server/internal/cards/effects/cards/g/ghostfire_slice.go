package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghostfire Slice — Instant {2}{R} (Modern Horizons 3, #5467):
//
//	"Devoid (This card has no color.)
//	 This spell costs {2} less to cast if an opponent controls a
//	 multicolored permanent.
//	 Ghostfire Slice deals 4 damage to any target."
//
// Devoid is the engine's colour-removing keyword (CR 702.114a). The
// reduction is Hagra Mauling's self cost modifier (CR 601.2f) over a
// permanent an opponent controls with two or more colours, read through
// the layers as the spell is priced; it takes off generic mana only, so
// the spell then costs {R}. The damage is Lightning Bolt's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "986a7ad0-527a-4079-aa1b-c8acdf087020",
		Name:            "Ghostfire Slice",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Targets:         effects.TargetAny(),
		Purpose:         game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 4)), Answers: game.AnswerRemove},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "This spell costs {2} less to cast if an opponent controls a multicolored permanent.",
				effects.PermanentsOnBattlefieldAtLeast(1, effects.And(effects.OpponentControls(), effects.Multicolored()))),
		},
		OnResolve: effects.DamageToFirstTargetStep(4),
	})
}
