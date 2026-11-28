package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Flotilla — Creature — Goblin {2}{R}, 2/2 (Fallen Empires, S82 OLD b854,
// #5637):
//
//	"Islandwalk
//	 At the beginning of each combat, unless you pay {R}, whenever this creature
//	 blocks or becomes blocked by a creature this combat, that creature gains
//	 first strike until end of turn."
//
// The beginning-of-combat trigger asks its controller to pay {R}; declining (or
// being unable) creates a delayed triggered ability (CR 603.7) that watches the
// block pairs of this combat only, so an additional combat starts fresh. Each
// pair names the other creature, which gains first strike as the trigger
// resolves.
//
// No simplification.
func init() {
	const label = "Goblin Flotilla — unless you pay {R}, whenever this creature blocks or becomes blocked by a creature this combat, that creature gains first strike until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "014ca38f-d4ac-4417-bfa8-785a7d6fce43",
		Name:            "Goblin Flotilla",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfEachCombat(label, func(g *game.Game, item *game.StackItem) error {
				return effects.PayUnless{
					Chooser:  item.Controller,
					Cost:     "{R}",
					Question: "Goblin Flotilla — pay {R}? If you don't, creatures it blocks or is blocked by gain first strike this combat",
					Else:     effects.S82OldB854FlotillaNotPaid,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
