package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lunar Rejection — Instant {1}{U}:
//
//	"Cleave {3}{U} (You may cast this spell for its cleave cost. If you
//	 do, remove the words in square brackets.)
//	 Return target [Wolf or Werewolf] creature to its owner's hand.
//	 Draw a card."
//
// Hard-cast it needs a Wolf or a Werewolf; cleaved any creature (CR
// 702.148a). The draw happens whether or not the bounce did, as long as
// the spell resolves at all.
func init() {
	Register(Spec{
		OracleID:     "d2f693d9-f338-41d6-b7e2-ddaa5f4bcc3b",
		Name:         "Lunar Rejection",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      TargetCreature("target Wolf or Werewolf creature", AnySubtype("Wolf", "Werewolf")),
		AlternativeCosts: []game.AlternativeCost{
			Cleave("{3}{U}", TargetCreature("target creature")),
		},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				t, ok := firstTargetCard(ctx.Item)
				if !ok {
					return nil
				}
				return BounceToHand{Target: t.ID}.Apply(ctx)
			}),
			DrawCards{N: 1},
		),
	})
}
