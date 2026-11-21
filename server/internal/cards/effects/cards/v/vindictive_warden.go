package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vindictive Warden — Creature — Human Soldier {2}{B/R}, 2/3 (Avatar: The
// Last Airbender, #4459):
//
//	"Menace
//	 Firebending 1 (Whenever this creature attacks, add {R}. This mana
//	 lasts until end of combat.)
//	 {3}: This creature deals 1 damage to each opponent."
//
// Firebending is Avatar Roku's (`firebending`); the ping is Merciless
// Enforcers' at a different cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "43b339ec-4d7d-41ac-a5ab-96c6a6de00e2",
		Name:            "Vindictive Warden",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Vindictive Warden — firebending 1", effects.Do(effects.NewFirebending(1))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: This creature deals 1 damage to each opponent.",
			Purpose: game.Purpose{Answers: game.AnswerValue, DamageEachOpponent: 1},
			Cost:    effects.ManaCost("{3}"),
			Effect:  effects.Do(effects.DamageEachOpponent{N: 1}),
		}},
	})
}
