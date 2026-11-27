package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Samite Elder — Creature — Human Cleric {2}{W}, 1/2 (Planeshift, S82 OLD
// #5615):
//
//	"{T}: Choose target permanent you control. Creatures you control gain
//	 protection from each of that permanent's colors until end of turn."
//
// The creatures you control are fixed as the ability resolves (CR 611.2c), and
// the colours are read from the chosen permanent then (CR 608.2h); a colourless
// permanent grants nothing. Protection from each colour is one protection
// quality per colour (CR 702.16).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8feb27aa-1f62-46e5-ba20-72c712de0862",
		Name:         "Samite Elder",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Choose target permanent you control. Creatures you control gain protection from each of that permanent's colors until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    game.AbilityCost{Tap: true},
			Targets: effects.TargetPermanent("target permanent you control", effects.YouControl()),
			Effect:  effects.OldB639ProtectionFromThatPermanentsColors("Samite Elder"),
		}},
	})
}
