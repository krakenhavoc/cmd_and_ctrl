package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tahngarth, Talruum Hero — Legendary Creature — Minotaur Warrior {3}{R}{R},
// 4/4 (Planeshift, S82 OLD #5615):
//
//	"Vigilance
//	 {1}{R}, {T}: Tahngarth deals damage equal to its power to target creature.
//	 That creature deals damage equal to its power to Tahngarth."
//
// Predatory Urge's mutual blows (S82OldB432MutualDamage): both amounts read as
// the ability resolves and dealt as one damage event, so deathtouch and
// lifelink on either side apply; a target that has left deals none. It is not a
// fight (CR 701.14 is not used). Vigilance is a printed keyword; the {T} keeps
// it from being activated the turn it arrives (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c5b4131e-13aa-4a22-b8f5-e5f9d366cb4f",
		Name:            "Tahngarth, Talruum Hero",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, {T}: Tahngarth deals damage equal to its power to target creature. That creature deals damage equal to its power to Tahngarth.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    game.AbilityCost{Mana: "{1}{R}", Tap: true},
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.S82OldB432MutualDamage,
		}},
	})
}
