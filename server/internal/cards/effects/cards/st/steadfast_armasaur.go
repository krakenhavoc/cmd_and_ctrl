package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steadfast Armasaur — Creature — Dinosaur {3}{W}, 2/3 (Ixalan, #5532):
//
//	"Vigilance
//	 {1}{W}, {T}: This creature deals damage equal to its toughness to
//	 target creature blocking or blocked by it."
//
// The target is a creature in combat with the Armasaur, judged as the
// ability goes on the stack and again as it resolves (CR 608.2b). The
// amount is the Armasaur's toughness as it resolves (its last-known
// toughness if it has left, CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dc0bd3a3-f34e-4d81-b9e8-b172b005d671",
		Name:            "Steadfast Armasaur",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, {T}: This creature deals damage equal to its toughness to target creature blocking or blocked by it.",
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Targets: effects.BlockingOrBlockedBySource(effects.TargetCreature("target creature blocking or blocked by this creature")),
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect:  effects.SteadfastArmasaurDamage,
		}},
	})
}
