package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archon of the Triumvirate — Creature — Archon {5}{W}{U}, 4/5 (Return to
// Ravnica, #5566):
//
//	"Flying
//	 Whenever this creature attacks, detain up to two target nonland permanents
//	 your opponents control."
//
// Up to two targets, chosen as the trigger goes on the stack; each one still legal
// on resolution is detained until your next turn (CR 701.35a): it can't attack or
// block and its activated abilities can't be activated. With no target chosen the
// trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "27ecb717-6251-4066-b57e-7070e9e44bbf",
		Name:            "Archon of the Triumvirate",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Archon of the Triumvirate — detain up to two target nonland permanents your opponents control",
				effects.OldB423DetainTargets("Archon of the Triumvirate — detain")),
				effects.TargetPermanent("up to two target nonland permanents your opponents control", effects.Nonland(), effects.OpponentControls()).WithCount(0, 2)),
		},
	})
}
