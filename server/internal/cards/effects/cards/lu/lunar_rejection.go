package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "d2f693d9-f338-41d6-b7e2-ddaa5f4bcc3b",
		Name:         "Lunar Rejection",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetCreature("target Wolf or Werewolf creature", effects.AnySubtype("Wolf", "Werewolf")),
		AlternativeCosts: []game.AlternativeCost{
			effects.Cleave("{3}{U}", effects.TargetCreature("target creature")),
		},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := effects.FirstTargetCard(ctx.Item)
				if !ok {
					return nil
				}
				return effects.BounceToHand{Target: t.ID}.Apply(ctx)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
