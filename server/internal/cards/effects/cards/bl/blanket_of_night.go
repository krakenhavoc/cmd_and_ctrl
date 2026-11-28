package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blanket of Night — Enchantment {1}{B}{B} (Visions, S82 OLD, #5631):
//
//	"Each land is a Swamp in addition to its other land types."
//
// Urborg, Tomb of Yawgmoth's one Layer-4 static, over every land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10e95489-a94d-4523-964c-ec9753103a62",
		Name:         "Blanket of Night",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.S82OldB788EachLandIsAlso("Swamp", "Each land is a Swamp in addition to its other land types."),
		},
	})
}
