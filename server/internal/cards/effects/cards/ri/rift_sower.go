package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rift Sower — Creature — Elf Druid {2}{G}, 1/3 (Modern Horizons 2):
//
//	"{T}: Add one mana of any color.
//	 Suspend 2—{G} (Rather than cast this card from your hand, you may pay {G} and
//	 exile it with two time counters on it. At the beginning of your upkeep, remove
//	 a time counter. When the last is removed, you may cast it without paying its
//	 mana cost. It has haste.)"
//
// A creature with a mana cost, so it can be cast normally or suspended. The
// suspend special action is the shared one: the free cast gives the creature
// haste (CR 702.62a), so it can tap for mana the turn it arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f9deeea1-bf80-4832-aac3-373fc1f4b41d",
		Name:         "Rift Sower",
		Completeness: effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{
			effects.Suspend(2, "{G}"),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
