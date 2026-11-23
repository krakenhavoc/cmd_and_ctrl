package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boros Locket — Artifact {3} (Guilds of Ravnica, #5525):
//
//	"{T}: Add {R} or {W}.
//	 {R/W}{R/W}{R/W}{R/W}, {T}, Sacrifice this artifact: Draw two cards."
//
// Azorius Locket's shape: a two-colour mana rock, and a draw-two whose cost
// is four hybrid symbols (each paid with {R} or {W}), the tap and the
// sacrifice, all paid at announcement.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b044b4fd-d504-4a11-80bf-eb8fb0b6e21e",
		Name:         "Boros Locket",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R|W}",
			Label:    "Add {R} or {W}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R/W}{R/W}{R/W}{R/W}, {T}, Sacrifice this artifact: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{R/W}{R/W}{R/W}{R/W}"},
			Effect:  effects.Do(effects.DrawCards{N: 2}),
		}},
	})
}
