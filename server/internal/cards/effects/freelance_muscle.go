package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Freelance Muscle — Creature — Rhino Warrior {4}{G}, 4/4 (Streets of New
// Capenna, S82):
//
//	"Whenever this creature attacks or blocks, it gets +X/+X until end of
//	 turn, where X is the greatest power and/or toughness among other
//	 creatures you control."
//
// Burning Sun Cavalry's attacks-or-blocks trigger with Tuya Bearclaw's
// count. X is the largest number among both the power and the toughness of
// every other creature its controller controls, counted once as the trigger
// resolves and then fixed. With none, or none above 0, X is 0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a749efc-e333-448e-8652-2dc4e6633e6b",
		Name:         "Freelance Muscle",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source) || selfBlocksOnce(ev, source)
				},
				"Freelance Muscle — it gets +X/+X until end of turn, X the greatest power and/or toughness among other creatures you control",
				freelanceMusclePump),
		},
	})
}

func freelanceMusclePump(g *game.Game, item *game.StackItem) error {
	x := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == item.SourceCardID || c.Controller != item.Controller || !c.IsCreature() {
			continue
		}
		if p := c.PowerForComparison(); p > x {
			x = p
		}
		if t := c.CurrentToughness(); t > x {
			x = t
		}
	}
	return thisGetsUntilEndOfTurn(x, x, fmt.Sprintf("Freelance Muscle — +%d/+%d until end of turn", x, x))(g, item)
}
