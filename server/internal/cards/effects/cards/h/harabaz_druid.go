package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Harabaz Druid — Creature — Human Druid Ally {1}{G}, 0/1 (Worldwake, S82 OLD b487):
//
//	"{T}: Add X mana of any one color, where X is the number of Allies you
//	 control."
//
// A mana ability (CR 605.1a) whose amount is counted as it is activated, the
// Druid included, all of one color the player picks (Accomplished Alchemist's
// second ability with the Allies you control for the amount).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ead985ec-f29f-4a3b-b8b1-061142cc5bd1",
		Name:         "Harabaz Druid",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{Tap: true},
			ProducedAmount: effects.ProducedOneColor(func(g *game.Game, controller, source uuid.UUID) int {
				src, _ := g.LookupCardForEffect(source)
				return effects.S82OldB487AlliesYouControl(g, controller, &src)
			}),
			Label: "Add X mana of any one color, where X is the number of Allies you control",
		}},
	})
}
