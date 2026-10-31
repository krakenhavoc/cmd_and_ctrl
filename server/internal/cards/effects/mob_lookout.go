package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mob Lookout — Creature — Human Rogue Villain {1}{U/B}, 0/3 (Marvel
// Super Heroes Commander, #4455):
//
//	"When this creature enters, target creature you control connives."
//
// The target is chosen as the trigger goes on the stack and may be Mob
// Lookout itself; the connive (ConniveEach) draws, discards, and puts
// the counter on the target if the discard was a nonland card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "47654b36-b215-4333-aa70-d88d627447e3",
		Name:         "Mob Lookout",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Mob Lookout — target creature you control connives", func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return ConniveEach{Controller: ctx.Controller(), Creatures: []uuid.UUID{id}}.Apply(ctx)
				}),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
