package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silumgar Monument — Artifact {3} (Dragons of Tarkir, S82 old batch 332):
//
//	"{T}: Add {U} or {B}.
//	 {4}{U}{B}: This artifact becomes a 4/4 blue and black Dragon artifact creature
//	 with flying until end of turn."
//
// Boros Keyrune's shape: a two-colour mana ability and an animation that adds the
// creature type and the Dragon subtype (layer 4), sets the colours (layer 5),
// grants flying (layer 6) and sets the base power and toughness (layer 7b)
// until end of turn (GtcKeyruneAnimate). It stays an artifact and keeps its mana
// ability, which the engine refuses while it is a summoning-sick creature
// (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b88268c-7fcb-4fd2-96cd-f93b37413283",
		Name:         "Silumgar Monument",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U|B}",
			Label:    "Add {U} or {B}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{U}{B}: This artifact becomes a 4/4 blue and black Dragon artifact creature with flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{4}{U}{B}"),
			Effect: effects.GtcKeyruneAnimate("Silumgar Monument — a 4/4 blue and black Dragon artifact creature with flying until end of turn",
				4, 4, "Dragon", []string{"U", "B"}, "flying"),
		}},
	})
}
