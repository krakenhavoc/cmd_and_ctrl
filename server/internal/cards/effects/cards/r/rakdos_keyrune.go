package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rakdos Keyrune — Artifact {3} (Return to Ravnica, #5566):
//
//	"{T}: Add {B} or {R}.
//	 {B}{R}: This artifact becomes a 3/1 black and red Devil artifact
//	 creature with first strike until end of turn."
//
// The Gatecrash Keyrunes' shape (gtcKeyruneAnimate): two colours of mana from
// one tap, and an animation that leaves it an artifact. A Keyrune that came
// under your control this turn can animate but can't attack or tap for mana
// while it is a creature (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3986b284-f943-4570-ba26-3419910feeef",
		Name:         "Rakdos Keyrune",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B|R}",
			Label:    "Add {B} or {R}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{R}: This artifact becomes a 3/1 black and red Devil artifact creature with first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{B}{R}"),
			Effect: effects.GtcKeyruneAnimate("Rakdos Keyrune — a 3/1 black and red Devil artifact creature with first strike until end of turn",
				3, 1, "Devil", []string{"B", "R"}, "first strike"),
		}},
	})
}
