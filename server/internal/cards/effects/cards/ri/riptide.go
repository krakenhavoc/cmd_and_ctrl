package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Riptide — Instant {U} (The Dark, #5638):
//
//	"Tap all blue creatures."
//
// The set is every blue creature on the battlefield as it resolves,
// whoever controls it, read through the creatures' current colours.
// Each is tapped as a member of the group, the way Cryptic Command
// taps all creatures your opponents control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b4541f38-23c1-4f51-a35b-0b222cdaed2c",
		Name:         "Riptide",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.OfColor("U"))) {
				if !c.Tapped {
					ids = append(ids, c.InstanceID)
				}
			}
			for _, id := range ids {
				if err := (effects.TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
