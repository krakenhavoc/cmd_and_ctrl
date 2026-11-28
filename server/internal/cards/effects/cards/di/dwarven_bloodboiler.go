package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwarven Bloodboiler — Creature — Dwarf {R}{R}{R}, 2/2 (Judgment):
//
//	"Tap an untapped Dwarf you control: Target creature gets +2/+0 until
//	 end of turn."
//
// Zombie Trailblazer's tap-an-untapped-<type> cost (not the {T} symbol, CR
// 302.6, so a Dwarf that arrived this turn may pay it, the Bloodboiler
// itself included) with a targeted pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cbbca098-0b33-4693-8ea7-a2005b79af3d",
		Name:         "Dwarven Bloodboiler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped Dwarf you control: Target creature gets +2/+0 until end of turn.",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0)), game.AnswerPump),
			Cost:    effects.TapAnUntapped("an untapped Dwarf you control", effects.HasSubtype("Dwarf")),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.BoostTargetUntilEOT("Dwarven Bloodboiler — +2/+0", 2, 0),
		}},
	})
}
