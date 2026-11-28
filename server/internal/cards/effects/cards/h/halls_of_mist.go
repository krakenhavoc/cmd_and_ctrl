package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Halls of Mist — Land (Ice Age, S82 OLD batch 838, #5636):
//
//	"Cumulative upkeep {1}
//	 Creatures that attacked during their controller's last turn can't attack."
//
// Giant Turtle's last-turn attack record (ADR 0108 §6) as a layer 6 static over
// every creature, any controller's: a creature that was declared as an attacker
// during its controller's last turn (the same object, CR 400.7) can't attack.
// The record changes only as a turn ends, where the layer cache is already
// invalidated as the next turn begins. Only declarations count (CR 508.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57d7deaf-ee08-49ab-b239-1c77a1f65225",
		Name:         "Halls of Mist",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.CumulativeUpkeep("Halls of Mist — cumulative upkeep {1}", "{1}")},
		Static: []game.StaticAbility{{
			Label: "Creatures that attacked during their controller's last turn can't attack.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, _ *game.Card) bool {
				return target.IsCreature() && g.AttackedDuringControllersLastTurn(target)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack
			},
		}},
	})
}
