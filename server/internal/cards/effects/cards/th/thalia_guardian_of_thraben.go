package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thalia, Guardian of Thraben — Legendary Creature — Human Soldier
// {1}{W}, 2/1:
//
//	"First strike
//	 Noncreature spells cost {1} more to cast."
//
// Thorn of Amethyst attached to a 2/1 first striker, which is why it
// sees far more play than the artifact: the tax comes with a clock,
// and the white weenie deck that plays it pays the tax least.
//
// Symmetrical as printed — Thalia's own controller pays it on their
// own removal spell too.
//
// First strike rides PrintedKeywords rather than a Static entry.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9b7f1d05-707c-4ed3-9f0e-8ced1232c2ee",
		Name:            "Thalia, Guardian of Thraben",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		CostModifiers: []game.CostModifier{
			effects.CostsMore(1, "Noncreature spells cost {1} more to cast.", effects.NoncreatureSpell()),
		},
	})
}
