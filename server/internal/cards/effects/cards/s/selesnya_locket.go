package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Selesnya Locket — Artifact {3} (Guilds of Ravnica, #5525):
//
//	"{T}: Add {G} or {W}.
//	 {G/W}{G/W}{G/W}{G/W}, {T}, Sacrifice this artifact: Draw two cards."
//
// Azorius Locket's shape: a two-colour mana rock, and a draw-two whose cost
// is four hybrid symbols (each paid with {G} or {W}), the tap and the
// sacrifice, all paid at announcement.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ecd98d5-8d66-4cb0-bada-23dfa0faacc0",
		Name:         "Selesnya Locket",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G|W}",
			Label:    "Add {G} or {W}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G/W}{G/W}{G/W}{G/W}, {T}, Sacrifice this artifact: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{G/W}{G/W}{G/W}{G/W}"},
			Effect:  effects.Do(effects.DrawCards{N: 2}),
		}},
	})
}
