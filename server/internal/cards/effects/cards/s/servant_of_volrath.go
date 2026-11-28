package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Servant of Volrath — Creature — Minion {2}{B}, 3/3 (Tempest, S82 OLD, #5628):
//
//	"When this creature leaves the battlefield, sacrifice a creature."
//
// The leave trigger fires however the Servant goes (dies, exile, bounce). As it
// resolves its controller sacrifices a creature of their choice, which is
// another one: the Servant has gone already. With no creature left, nothing
// is sacrificed.
//
// No simplification.
func init() {
	const name = "Servant of Volrath"
	effects.Register(effects.Spec{
		OracleID:     "4e389b61-7b31-42ba-a19b-2502f63fd34a",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisLeaves(name+" — sacrifice a creature", func(g *game.Game, item *game.StackItem) error {
				g.PlayerSacrificesNForEffect(item.SourceCardID, item.Controller, effects.SacrificeSpec("a creature", effects.Creature()), name+" — sacrifice a creature", 1)
				return nil
			}),
		},
	})
}
