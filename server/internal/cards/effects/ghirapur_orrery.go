package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:                         "c6bb3a58-131a-4d7b-8c07-173a7de23b06",
		Name:                             "Ghirapur Orrery",
		Completeness:                     CompletenessFull,
		Purpose:                          game.Purpose{ExtraLandDrops: 1},
		AdditionalLandPlays:              1,
		AdditionalLandPlaysForEachPlayer: true,
		Triggered: []game.TriggeredAbility{
			InterveningIf(cmdr2B01CondActorHandEmpty{},
				AtEachUpkeep("Ghirapur Orrery — that player draws three cards",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Trigger.Event.Actor, N: 3}.Apply(NewContext(g, item))
					})),
		},
	})
}
