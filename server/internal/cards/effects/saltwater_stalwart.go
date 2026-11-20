package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "04472c4f-0df3-4de8-9cc2-882d95b822be",
		Name:         "Saltwater Stalwart",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisDealsDamageToAnOpponent("Saltwater Stalwart — target player draws a card", theTargetPlayerDraws(1)),
				TargetPlayer("target player")),
		},
	})
}
