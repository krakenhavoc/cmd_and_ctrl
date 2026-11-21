package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Martyred Rusalka — Creature — Spirit {W}, 1/1 (Guildpact, #5599):
//
//	"{W}, Sacrifice a creature: Target creature can't attack this turn."
//
// Netter en-Dal's restriction (nemTargetRestrictedThisTurn with
// game.CantAttack) behind High Market's sacrifice: any creature its
// controller controls pays the cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21bf5ae1-d7a0-406d-88fd-009dfa919096",
		Name:         "Martyred Rusalka",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, Sacrifice a creature: Target creature can't attack this turn.",
			Purpose: game.Purpose{Targets: game.ForTargets(effects.HarmUntilEOT(0, 0, 0)), Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.SacrificeACreature()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.NemTargetRestrictedThisTurn(game.CantAttack, "Martyred Rusalka — can't attack this turn"),
		}},
	})
}
