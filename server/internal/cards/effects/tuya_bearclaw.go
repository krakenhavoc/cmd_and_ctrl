package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tuya Bearclaw — Legendary Creature — Human Warrior {1}{R}{G}, 2/2
// (Commander Legends, #5507):
//
//	"Whenever Tuya Bearclaw attacks, it gets +X/+X until end of turn,
//	 where X is the greatest power among other creatures you control."
//
// X is counted once, as the trigger resolves, over the creatures the
// trigger's controller controls other than Tuya, and the bonus is then
// fixed (the rulings). With no other creature, or when the greatest power
// among them is 0 or less, X is 0 and nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "63e8ac70-0ec3-4bef-832f-e3f4c8a692ad",
		Name:         "Tuya Bearclaw",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Tuya Bearclaw — it gets +X/+X until end of turn, X the greatest power among other creatures you control", tuyaBearclawPump),
		},
	})
}

// tuyaBearclawPump reads X and pumps Tuya by it.
func tuyaBearclawPump(g *game.Game, item *game.StackItem) error {
	x := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == item.SourceCardID || c.Controller != item.Controller || !c.IsCreature() {
			continue
		}
		if p := c.PowerForComparison(); p > x {
			x = p
		}
	}
	return thisGetsUntilEndOfTurn(x, x, fmt.Sprintf("Tuya Bearclaw — +%d/+%d until end of turn", x, x))(g, item)
}
