package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Atarka Monument — Artifact {3} (Commander Starter Decks):
//
//	"{T}: Add {R} or {G}.
//	 {4}{R}{G}: This artifact becomes a 4/4 red and green Dragon artifact
//	 creature with flying until end of turn."
//
// Boros Keyrune's shape: the same two-colour mana ability and the shared
// animation (GtcKeyruneAnimate), still an artifact, with flying. A Monument
// that came under your control this turn can animate but can't attack or
// tap for mana while it is a creature (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb287c04-97f5-4e47-beb1-66cc7d15e748",
		Name:         "Atarka Monument",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R|G}",
			Label:    "Add {R} or {G}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{R}{G}: This artifact becomes a 4/4 red and green Dragon artifact creature with flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{4}{R}{G}"),
			Effect: effects.GtcKeyruneAnimate("Atarka Monument — a 4/4 red and green Dragon artifact creature with flying until end of turn",
				4, 4, "Dragon", []string{"R", "G"}, "flying"),
		}},
	})
}
