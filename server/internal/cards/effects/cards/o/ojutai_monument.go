package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ojutai Monument — Artifact {3} (Dragons of Tarkir, S82 old batch 332):
//
//	"{T}: Add {W} or {U}.
//	 {4}{W}{U}: This artifact becomes a 4/4 white and blue Dragon artifact creature
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
		OracleID:     "d6b4810f-2bee-4c83-b13d-c46fb7dd8e08",
		Name:         "Ojutai Monument",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U}",
			Label:    "Add {W} or {U}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{W}{U}: This artifact becomes a 4/4 white and blue Dragon artifact creature with flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{4}{W}{U}"),
			Effect: effects.GtcKeyruneAnimate("Ojutai Monument — a 4/4 white and blue Dragon artifact creature with flying until end of turn",
				4, 4, "Dragon", []string{"W", "U"}, "flying"),
		}},
	})
}
