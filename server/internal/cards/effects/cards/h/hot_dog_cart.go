package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hot Dog Cart — Artifact {3}:
//
//	"When this artifact enters, create a Food token. (It's an artifact
//	 with "{2}, {T}, Sacrifice this token: You gain 3 life.")
//	 {T}: Add one mana of any color."
//
// An enters trigger that makes one Food, and a mana ability (no target,
// produces mana, so it skips the stack). No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf27f1f8-2d95-4a00-97f8-72a244837f08",
		Name:         "Hot Dog Cart",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hot Dog Cart — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
	})
}
