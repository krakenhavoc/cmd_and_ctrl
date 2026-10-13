package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kuon, Ogre Ascendant // Kuon's Essence — {B}{B}{B} Legendary Creature
// — Ogre Monk 2/4 // Legendary Enchantment, a flip card (#3216, ADR
// 0151):
//
//	Top:     "At the beginning of the end step, if three or more
//	          creatures died this turn, flip Kuon."
//	Flipped: "At the beginning of each player's upkeep, that player
//	          sacrifices a creature of their choice."
//
// "The end step" is every turn's, and the count is the table-wide tally
// of creatures that died (Bloodline Recollector's, b11CreaturesDied
// ThisTurn), checked as the end step begins and again as the ability
// resolves (CR 603.4). The Essence is Mana Vortex's upkeep edict: the
// player whose upkeep it is chooses and sacrifices, Kuon's controller
// included. It is black and {B}{B}{B} (CR 710.1c).
//
// No simplification.
const kuonOracleID = "f1b1cf99-b44c-4bcc-b0ed-98f69e66c54a"

func init() {
	threeDied := func(g *game.Game, _ uuid.UUID) bool { return b11CreaturesDiedThisTurn(g) >= 3 }
	Register(Spec{
		OracleID:     kuonOracleID,
		Name:         "Kuon, Ogre Ascendant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return threeDied(g, uuid.Nil)
			}, "Kuon, Ogre Ascendant — flip it", flipThisIf(threeDied)),
		},
	})
	Register(Spec{
		OracleID:     kuonOracleID + "#1",
		Name:         "Kuon's Essence",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Kuon's Essence — that player sacrifices a creature", func(g *game.Game, item *game.StackItem) error {
				player := NewContext(g, item).Trigger().Event.Actor
				if player == uuid.Nil {
					return nil
				}
				g.PlayerSacrificesForEffect(item.SourceCardID, player, sacrificeSpec("a creature", Creature()),
					"Kuon's Essence — sacrifice a creature")
				return nil
			}),
		},
	})
}
