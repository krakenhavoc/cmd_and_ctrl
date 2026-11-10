package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Overgrown Zealot — Creature — Elf Druid {1}{G}, 0/4 (Duskmourn: House of
// Horror, #4466):
//
//	"{T}: Add one mana of any color.
//	 {T}: Add two mana of any one color. Spend this mana only to turn
//	 permanents face up."
//
// Two mana abilities. The first is Birds of Paradise's. The second is
// Sanctum Weaver's "N mana of any one color" (ProducedOneColor, a fixed
// two) with Creeping Peeper's restriction (ManaRestrictTurnFaceUp), so the
// mana pays a morph, megamorph or disguise cost, or the turn-face-up cost
// of a manifested or cloaked permanent, and nothing else (CR 116.2g).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5146b291-1170-481f-a4b9-709dd5c4b09c",
		Name:         "Overgrown Zealot",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			},
			{
				Cost:           ManaAbilityCost{Tap: true},
				ProducedAmount: ProducedOneColor(func(*game.Game, uuid.UUID, uuid.UUID) int { return 2 }),
				Label:          "Add two mana of any one color. Spend this mana only to turn permanents face up",
				Restrictions:   []string{game.ManaRestrictTurnFaceUp},
			},
		},
	})
}
