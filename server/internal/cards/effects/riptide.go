package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "b4541f38-23c1-4f51-a35b-0b222cdaed2c",
		Name:         "Riptide",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, c := range MatchingBattlefield(ctx, And(Creature(), OfColor("U"))) {
				if !c.Tapped {
					ids = append(ids, c.InstanceID)
				}
			}
			for _, id := range ids {
				if err := (TapTarget{Target: id}).Apply(ctx.asGroupMember()); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
