package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lieutenant Kirtar — Legendary Creature — Bird Soldier {1}{W}{W}, 2/2
// (Dominaria Remastered, #5948):
//
//	"Flying
//	 {1}{W}, Sacrifice Lieutenant Kirtar: Exile target attacking creature."
//
// The sacrifice is a cost, so the ability is on the stack when Kirtar has
// gone; the exile still happens if the target is still an attacking creature
// as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ad39e859-e4a6-49cc-b738-4578a5250538",
		Name:            "Lieutenant Kirtar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, Sacrifice Lieutenant Kirtar: Exile target attacking creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target attacking creature", effects.IsAttackingCreature()),
			Effect:  effects.ExileFirstLegalCardTarget,
		}},
	})
}
