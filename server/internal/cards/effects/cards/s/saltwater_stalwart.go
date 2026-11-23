package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saltwater Stalwart — Creature — Merfolk Warrior {3}{U}, 2/4 (Battlebond,
// #5529):
//
//	"Whenever this creature deals damage to an opponent, target player
//	 draws a card."
//
// Any damage, combat or not, dealt by the Stalwart to one of its
// controller's opponents (WheneverThisDealsDamageToAnOpponent, Looter
// il-Kor's trigger). The target is any player, chosen as the trigger goes
// on the stack; a player who has left the game draws nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04472c4f-0df3-4de8-9cc2-882d95b822be",
		Name:         "Saltwater Stalwart",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisDealsDamageToAnOpponent("Saltwater Stalwart — target player draws a card", effects.TheTargetPlayerDraws(1)),
				effects.TargetPlayer("target player")),
		},
	})
}
