package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loaming Shaman — Creature — Centaur Shaman {2}{G}, 3/2 (Ravnica
// Remastered, #5948):
//
//	"When this creature enters, target player shuffles any number of target
//	 cards from their graveyard into their library."
//
// Krosan Reclamation's two-clause target group (a player and any number of
// cards from THAT player's graveyard, FromTargetPlayersGraveyard) and its
// TargetPlayerShufflesTheirCards body: an illegal player does nothing, a legal
// one shuffles even when no card was picked, and a card that left the
// graveyard in the meantime stays where it is (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf7352bf-3971-4bb7-88b9-65ab9a219289",
		Name:         "Loaming Shaman",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Loaming Shaman — target player shuffles any number of target cards from their graveyard into their library",
					func(g *game.Game, item *game.StackItem) error {
						return effects.TargetPlayerShufflesTheirCards(effects.NewContext(g, item), item.Targets, 0, 1)
					}),
				effects.Clauses(
					effects.TargetPlayer("target player"),
					effects.FromTargetPlayersGraveyard(
						effects.TargetCardInGraveyard("any number of target cards from their graveyard").WithCount(0, 0), 0),
				)),
		},
	})
}
