package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Karfell Harbinger — Creature — Zombie Wizard {1}{U}, 1/3:
//
//	"{T}: Add {U}. Spend this mana only to foretell a card from your hand or
//	 cast an instant or sorcery spell."
//
// The restriction is an OR of two clauses, one tag (ManaRestrictAnyOf): the
// cost of the foretell special action (CR 702.143a; a special action is
// neither a cast nor an activation, so it has its own spend purpose,
// game.ManaRestrictForetell), or the cost of casting an instant or a sorcery
// spell. Nothing else — a creature spell, an ability — can use it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0507a9d2-cbad-4a7b-b888-fbd5ac1ed02f",
		Name:         "Karfell Harbinger",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}. Spend this mana only to foretell a card from your hand or cast an instant or sorcery spell",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictForetell},
				[]string{effects.ManaRestrictCast, effects.ManaRestrictAnyType("Instant", "Sorcery")},
			)},
		}},
	})
}
