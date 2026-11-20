package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Purple Dragon Punks — {1}{R} Creature — Human Rogue 2/2 (Teenage Mutant
// Ninja Turtles, #4457):
//
//	"{T}: Add {R}. Spend this mana only to cast an artifact spell or to
//	 activate an ability."
//
// Gwenna, Eyes of Gaea's one ManaRestrictAnyOf tag with different
// alternatives (CR 106.6): a cast whose spell is an artifact, or any
// activation (the source of the ability does not matter, and a mana
// ability with a mana cost is an activation too). Anything else, a
// nonartifact spell or a cost that is neither, can't spend it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1edb950-a118-4b4d-a0ef-9a6683d74c9f",
		Name:         "Purple Dragon Punks",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "{T}: Add {R}. Spend this mana only to cast an artifact spell or to activate an ability.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Artifact")},
				[]string{game.ManaRestrictActivate},
			)},
		}},
	})
}
