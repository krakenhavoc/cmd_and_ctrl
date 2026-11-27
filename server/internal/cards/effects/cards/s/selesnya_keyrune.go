package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Selesnya Keyrune — Artifact {3} (Return to Ravnica, #5566):
//
//	"{T}: Add {G} or {W}.
//	 {G}{W}: This artifact becomes a 3/3 green and white Wolf artifact
//	 creature until end of turn."
//
// The Gatecrash Keyrunes' shape (gtcKeyruneAnimate): two colours of mana from
// one tap, and an animation that leaves it an artifact. A Keyrune that came
// under your control this turn can animate but can't attack or tap for mana
// while it is a creature (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d4ecee6d-b23f-48e7-b7e5-347027225709",
		Name:         "Selesnya Keyrune",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G|W}",
			Label:    "Add {G} or {W}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}{W}: This artifact becomes a 3/3 green and white Wolf artifact creature until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{G}{W}"),
			Effect: effects.GtcKeyruneAnimate("Selesnya Keyrune — a 3/3 green and white Wolf artifact creature until end of turn",
				3, 3, "Wolf", []string{"G", "W"}, ""),
		}},
	})
}
