package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Golgari Keyrune — Artifact {3} (Return to Ravnica, #5566):
//
//	"{T}: Add {B} or {G}.
//	 {B}{G}: This artifact becomes a 2/2 black and green Insect artifact
//	 creature with deathtouch until end of turn."
//
// The Gatecrash Keyrunes' shape (gtcKeyruneAnimate): two colours of mana from
// one tap, and an animation that leaves it an artifact. A Keyrune that came
// under your control this turn can animate but can't attack or tap for mana
// while it is a creature (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec970eba-230a-4c41-874a-4a74abf3214e",
		Name:         "Golgari Keyrune",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B|G}",
			Label:    "Add {B} or {G}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{G}: This artifact becomes a 2/2 black and green Insect artifact creature with deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{B}{G}"),
			Effect: effects.GtcKeyruneAnimate("Golgari Keyrune — a 2/2 black and green Insect artifact creature with deathtouch until end of turn",
				2, 2, "Insect", []string{"B", "G"}, "deathtouch"),
		}},
	})
}
