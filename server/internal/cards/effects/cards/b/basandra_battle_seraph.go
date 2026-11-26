package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Basandra, Battle Seraph — Legendary Creature — Angel {3}{R}{W}, 4/4
// (Conspiracy, S82 OLD batch 363, #5557):
//
//	"Flying
//	 Players can't cast spells during combat.
//	 {R}: Target creature attacks this turn if able."
//
// The cast restriction is Marisi's, for every player rather than only the
// opponents: every step from the beginning of combat to the end of combat.
// The activated ability is Imps' Taunt's attack requirement
// (S82OldB363TargetAttacksThisTurnIfAble): it binds the creature's controller
// only if the creature can attack this turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4a480736-4cdc-4e3a-a8cb-d1c726982f9f",
		Name:            "Basandra, Battle Seraph",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CastRestrictions: []game.CastRestriction{{
			Label: "Basandra, Battle Seraph — players can't cast spells during combat.",
			Forbids: func(q game.CastQuery) bool {
				return effects.InCombatStep(q.Game.Turn.Step)
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Target creature attacks this turn if able.",
			Cost:    effects.ManaCost("{R}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.S82OldB363TargetAttacksThisTurnIfAble("Basandra, Battle Seraph"),
		}},
	})
}
