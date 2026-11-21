package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghirapur Orrery — Artifact {4}:
//
//	"Each player may play an additional land on each of their turns.
//	 At the beginning of each player's upkeep, if that player has no
//	 cards in hand, that player draws three cards."
//
// The land clause is Rites of Flourishing's (a table-wide extra land play,
// CR 305.2). The draw is a trigger on every player's upkeep with an
// intervening "if" (CR 603.4) read for that player, as it triggers and as
// it resolves; the cards go to that player, not to the Orrery's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                         "c6bb3a58-131a-4d7b-8c07-173a7de23b06",
		Name:                             "Ghirapur Orrery",
		Completeness:                     effects.CompletenessFull,
		Purpose:                          game.Purpose{ExtraLandDrops: 1},
		AdditionalLandPlays:              1,
		AdditionalLandPlaysForEachPlayer: true,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.Cmdr2B01CondActorHandEmpty{},
				effects.AtEachUpkeep("Ghirapur Orrery — that player draws three cards",
					func(g *game.Game, item *game.StackItem) error {
						return effects.DrawCards{Player: item.Trigger.Event.Actor, N: 3}.Apply(effects.NewContext(g, item))
					})),
		},
	})
}
