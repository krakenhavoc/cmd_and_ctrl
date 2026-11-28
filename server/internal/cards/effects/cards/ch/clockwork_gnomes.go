package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Clockwork Gnomes — Artifact Creature — Gnome {4}, 2/2 (Homelands, S82 OLD b821,
// #5635):
//
//	"{3}, {T}: Regenerate target artifact creature."
//
// Jhoira's Toolbox's regeneration shield (CR 701.19a) behind a {3} and tap
// cost. The Gnomes may regenerate themselves: they are an artifact creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21d269f2-c11c-4094-b526-500f0785f483",
		Name:         "Clockwork Gnomes",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Regenerate target artifact creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetCreature("target artifact creature", effects.Artifact()),
			Effect:  effects.RegenerateTheTargetPermanent,
		}},
	})
}
