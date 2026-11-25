package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glacial Grasp — Instant {2}{U} (Zendikar Rising, S82 OLD #5508):
//
//	"Tap target creature. Its controller mills two cards. That creature
//	 doesn't untap during its controller's next untap step.
//	 Draw a card."
//
// The creature's controller is read before anything happens to it. The
// freeze is Berg Strider's TapAndFreeze: it taps the creature and holds it
// through its controller's next untap step. The spell does nothing, the
// draw included, if its one target is illegal as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3006be3-a796-4238-95dd-1e8e47c063b5",
		Name:         "Glacial Grasp",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, found := ctx.Game.LookupCardForEffect(id)
			if err := (effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Glacial Grasp — it doesn't untap during its controller's next untap step"}).Apply(ctx); err != nil {
				return err
			}
			if found {
				if err := (effects.MillCards{Player: c.Controller, N: 2}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
