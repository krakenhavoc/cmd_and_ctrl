package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smolder Initiate — Creature — Elemental Shaman {B}, 1/1 (Shadowmoor, S82
// OLD b529, #5591):
//
//	"Whenever a player casts a black spell, you may pay {1}. If you do, target
//	 player loses 1 life."
//
// Apothecary Initiate's trigger with a target: the player is chosen as the
// ability goes on the stack, the {1} is asked as it resolves, and the life is
// lost only once it is paid and the target is still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0300d8a1-ff4e-4e05-a097-4822f80c47e1",
		Name:         "Smolder Initiate",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.OldB529WhenAPlayerCasts("B", "Smolder Initiate — you may pay {1}; if you do, target player loses 1 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return effects.OldB529MayPayOne("Smolder Initiate — pay {1} to make target player lose 1 life?",
								smolderInitiatePaid, t.ID)(g, item)
						}
					}
					return nil
				}),
				effects.TargetPlayer("target player")),
		},
	})
}

// smolderInitiatePaid is the paid branch; the target is the prompt's own carry. An on-disk key: never renamed, never reused.
var smolderInitiatePaid = effects.PayUnlessBranches("pay-unless/smolder-initiate-lose-1",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) < 1 {
			return nil
		}
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.Carry[0], -1)
	}, nil)
