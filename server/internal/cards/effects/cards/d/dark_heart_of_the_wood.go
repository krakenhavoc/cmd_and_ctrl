package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dark Heart of the Wood — Enchantment {B}{G} (Ravnica: City of Guilds, S82
// OLD b582):
//
//	"Sacrifice a Forest: You gain 3 life."
//
// Lake of the Dead's sacrifice-a-land-by-type cost (a Forest by subtype, so a
// Forest-typed nonbasic pays it), on an activated ability with no mana cost
// and no limit; the life is the controller's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c44f40da-867e-4237-b4b1-ed6feb1f37b7",
		Name:         "Dark Heart of the Wood",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a Forest: You gain 3 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Forest", effects.HasSubtype("Forest"))},
			Effect:  effects.Do(effects.GainLife{Amount: 3}),
		}},
	})
}
