package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ezekiel Sims, Spider-Totem — Legendary Creature — Spider Human Advisor
// {4}{G}, 3/5:
//
//	"Reach
//	 At the beginning of combat on your turn, target Spider you control
//	 gets +2/+2 until end of turn."
//
// The target is any Spider the controller has, Ezekiel himself included
// (he is a Spider); with none the trigger is removed (CR 603.3d).
// Arahbo's beginning-of-combat shape without "another". No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4074e897-82a4-4f57-86c2-4aab3f29ef9b",
		Name:            "Ezekiel Sims, Spider-Totem",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtBeginningOfYourCombat(
				"Ezekiel Sims, Spider-Totem — target Spider you control gets +2/+2",
				func(g *game.Game, item *game.StackItem) error {
					return effects.VsPumpTheTarget(effects.NewContext(g, item), 2, 2, "Ezekiel Sims, Spider-Totem — +2/+2")
				}),
				effects.TargetCreature("target Spider you control", effects.OfCreatureType("Spider"), effects.YouControl())),
		},
	})
}
