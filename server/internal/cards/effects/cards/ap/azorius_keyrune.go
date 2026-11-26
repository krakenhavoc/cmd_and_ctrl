package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Azorius Keyrune — Artifact {3} (Commander 2013, S82 OLD b387):
//
//	"{T}: Add {W} or {U}.
//	 {W}{U}: This artifact becomes a 2/2 white and blue Bird artifact creature
//	 with flying until end of turn."
//
// Boros Keyrune's mana ability and animation (GtcKeyruneAnimate) with the Bird's
// colours and flying. It is still an artifact; a Keyrune that came under your
// control this turn can animate but can't attack or tap for mana while it is a
// creature (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7266b491-54e6-4393-a448-d5ae99d965c6",
		Name:         "Azorius Keyrune",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U}",
			Label:    "Add {W} or {U}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{U}: This artifact becomes a 2/2 white and blue Bird artifact creature with flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{W}{U}"),
			Effect: effects.GtcKeyruneAnimate("Azorius Keyrune — a 2/2 white and blue Bird artifact creature with flying until end of turn",
				2, 2, "Bird", []string{"W", "U"}, "flying"),
		}},
	})
}
