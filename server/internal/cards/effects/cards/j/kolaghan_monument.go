package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kolaghan Monument — Artifact {3} (Dragons of Tarkir, S82 old batch 332):
//
//	"{T}: Add {B} or {R}.
//	 {4}{B}{R}: This artifact becomes a 4/4 black and red Dragon artifact creature
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
		OracleID:     "d8b79fed-9817-4b91-ac2e-e9fa955dd43c",
		Name:         "Kolaghan Monument",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B|R}",
			Label:    "Add {B} or {R}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{B}{R}: This artifact becomes a 4/4 black and red Dragon artifact creature with flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{4}{B}{R}"),
			Effect: effects.GtcKeyruneAnimate("Kolaghan Monument — a 4/4 black and red Dragon artifact creature with flying until end of turn",
				4, 4, "Dragon", []string{"B", "R"}, "flying"),
		}},
	})
}
