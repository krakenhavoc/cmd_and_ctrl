package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lapis Orb of Dragonkind — Artifact {2}{U} (Commander Legends: Battle for
// Baldur's Gate, #5948):
//
//	"{T}: Add {U}. When you spend this mana to cast a Dragon creature spell,
//	 scry 2."
//
// A mana ability with a spend rider (Scaled Nurturer's shape): the {U} is
// tagged to trigger once when it pays for a Dragon creature spell, and the
// trigger scries 2 as an ability of its own. Mana that pays for
// anything else triggers nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "c61ab8e1-ef23-465d-bcd3-f771434988b2",
		Name:           "Lapis Orb of Dragonkind",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Scry: 2},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
			SpendRiders: []game.ManaSpendRider{effects.WhenManaSpent("Lapis Orb of Dragonkind", game.ManaSpendTrigger{
				Label:  "Lapis Orb of Dragonkind — scry 2",
				Effect: effects.Do(effects.Scry{N: 2}),
			}, effects.ManaRestrictCast, effects.ManaRestrictType("Creature"), effects.ManaRestrictSubtype("Dragon"))},
		}},
	})
}
