package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Last One Standing — Sorcery {1}{B}{R} (Battlebond, #5529):
//
//	"Choose a creature at random, then destroy the rest."
//
// The random pick is made as the spell resolves, among every creature on
// the battlefield, from the keyed random stream (ChooseAtRandomForEffect),
// so an undo replays it. Every other creature is then destroyed at once
// (DestroyAllMatching, so indestructible and regeneration apply as for any
// wrath). With no creature on the battlefield nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b01465ea-cab9-4a59-81a6-8db4648b89f4",
		Name:         "Last One Standing",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var creatures []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCreature() {
					creatures = append(creatures, c.InstanceID)
				}
			}
			pick := ctx.Game.ChooseAtRandomForEffect(effects.RandomDraw(ctx), creatures, 1)
			if len(pick) != 1 {
				return nil
			}
			survivor := pick[0]
			return effects.DestroyAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return c.IsCreature() && c.InstanceID != survivor
			}}.Apply(ctx)
		},
	})
}
