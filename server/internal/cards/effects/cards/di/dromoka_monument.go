package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dromoka Monument — Artifact {3} (Dragons of Tarkir, S82 old batch 332):
//
//	"{T}: Add {G} or {W}.
//	 {4}{G}{W}: This artifact becomes a 4/4 green and white Dragon artifact creature
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
		OracleID:     "1c587200-02cc-4c5f-86fe-d13cc7bdad8b",
		Name:         "Dromoka Monument",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G|W}",
			Label:    "Add {G} or {W}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G}{W}: This artifact becomes a 4/4 green and white Dragon artifact creature with flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{4}{G}{W}"),
			Effect: effects.GtcKeyruneAnimate("Dromoka Monument — a 4/4 green and white Dragon artifact creature with flying until end of turn",
				4, 4, "Dragon", []string{"G", "W"}, "flying"),
		}},
	})
}
