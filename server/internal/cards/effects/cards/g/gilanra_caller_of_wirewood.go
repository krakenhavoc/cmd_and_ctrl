package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gilanra, Caller of Wirewood — Legendary Creature — Elf Druid {2}{G}, 1/2
// (Commander Legends, #5507):
//
//	"{T}: Add {G}. When you spend this mana to cast a spell with mana value 6
//	 or greater, draw a card.
//	 Partner"
//
// The mana ability carries Lapis Orb of Dragonkind's spend rider: the draw
// is a triggered ability that goes on the stack above the spell the mana paid
// for, and only when that mana was spent on a cast of a spell with mana value
// 6 or greater (a copy is not cast; an ability's cost does not count). Partner
// is a deck-construction rule.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "4bda947d-493f-47a0-b0d2-516a6ef6b2c0",
		Name:           "Gilanra, Caller of Wirewood",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Draws: 1},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
			SpendRiders: []game.ManaSpendRider{effects.WhenManaSpent("Gilanra, Caller of Wirewood", game.ManaSpendTrigger{
				Label:  "Gilanra, Caller of Wirewood — draw a card",
				Effect: effects.Do(effects.DrawCards{N: 1}),
			}, effects.ManaRestrictCast, game.ManaRestrictManaValueAtLeast(6))},
		}},
	})
}
