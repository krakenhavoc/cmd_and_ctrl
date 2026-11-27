package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lotus Guardian (Invasion, S82 OLD batch 645, #5616):
//
//	Artifact Creature — Dragon {7}, 4/4:
//	"Flying
//	{T}: Add one mana of any color."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "143d7a32-c15d-4a1f-ad81-cf1e6aa92bf3",
		Name:            "Lotus Guardian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ManaAbilities: []effects.ManaAbility{
			effects.SamiAnyColorMana(effects.ManaAbilityCost{Tap: true}, "{T}: Add one mana of any color."),
		},
	})
}
