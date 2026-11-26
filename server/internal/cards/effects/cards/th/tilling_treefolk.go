package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tilling Treefolk — {2}{G} Creature — Treefolk Druid, 1/3:
//
//	"When this creature enters, you may return up to two target land cards from
//	 your graveyard to your hand."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8a011d0d-da29-4287-b50c-12746abfa38c",
		Name:         "Tilling Treefolk",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Tilling Treefolk — return up to two target land cards from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
				}),
				effects.TargetCardInGraveyard("up to two target land cards from your graveyard", effects.YouOwn(), effects.Land()).WithCount(0, 2)),
				"Tilling Treefolk — return the lands to your hand?"),
		},
	})
}
