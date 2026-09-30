package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Optimistic Scavenger — Creature — Human Scout {W}, 1/1:
//
//	"Eerie — Whenever an enchantment you control enters and whenever you
//	fully unlock a Room, put a +1/+1 counter on target creature."
//
// Any creature may be the target, yours or not, as printed.
//
// Eerie is one ability with two conditions (Eerie, rooms.go): an
// enchantment entering under your control, or you fully unlocking a Room.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0180c67b-c5c8-4a67-9562-d51f9e7ffff5",
		Name:         "Optimistic Scavenger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(Eerie("Optimistic Scavenger — put a +1/+1 counter on target creature (eerie)", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}
				return nil
			}), TargetCreature("target creature")),
		},
	})
}
