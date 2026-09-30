package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scrabbling Skullcrab — Creature — Crab Skeleton {U}, 0/3:
//
//	"Eerie — Whenever an enchantment you control enters and whenever you
//	fully unlock a Room, target player mills two cards."
//
// The player is a target of the trigger, chosen as it goes on the stack.
//
// Eerie is one ability with two conditions (Eerie, rooms.go): an
// enchantment entering under your control, or you fully unlocking a Room.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1448489c-dfb8-43aa-ae3d-ed832381c1d9",
		Name:         "Scrabbling Skullcrab",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(Eerie("Scrabbling Skullcrab — target player mills two cards (eerie)", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				p := targetedPlayerOf(ctx)
				if p == uuid.Nil {
					return nil
				}
				return MillCards{Player: p, N: 2}.Apply(ctx)
			}), TargetPlayer("target player")),
		},
	})
}
