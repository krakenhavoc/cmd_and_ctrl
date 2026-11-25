package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phase Dolphin — Creature — Elemental Whale {2}{U}, 1/4 (Ikoria: Lair of
// Behemoths, #5512):
//
//	"Whenever this creature attacks, another target attacking creature
//	 can't be blocked this turn."
//
// The target is chosen as the trigger goes on the stack, among the other
// attacking creatures (any controller's, as printed); the restriction is
// applied as it resolves if the target is still legal (CR 608.2b), and
// lasts to end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b70b739b-b147-4abe-8e48-26e7ea02a9f6",
		Name:         "Phase Dolphin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Phase Dolphin — another target attacking creature can't be blocked this turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: "Phase Dolphin — can't be blocked"}.Apply(ctx)
					}),
				effects.Another(effects.TargetCreature("another target attacking creature", effects.AttackingCreature()))),
		},
	})
}
