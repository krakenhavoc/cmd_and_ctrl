package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Touch of the Void — Sorcery {2}{R}:
//
//	Devoid (This card has no color.)
//	Touch of the Void deals 3 damage to any target. If a creature dealt damage this way would die this turn, exile it instead.
//
// "A creature dealt damage this way" is registered from the damage's continuation, on a creature that was dealt more than 0 damage (ADR 0108 §1 decision 3).
//
// Devoid is declared in PrintedKeywords, and the engine reads it as CR
// 702.114a's colour-defining ability (#2152), so the card is colourless
// in every zone.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "6b530534-5c02-4874-9256-501102ef8a5f",
		Name:            "Touch of the Void",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Targets:         TargetAny(),
		Purpose:         ForTargets(DamageToTarget(0, 3)),
		OnResolve:       damageAnyTargetExileIfDealtDies(fixedAmount(3), false),
	})
}
