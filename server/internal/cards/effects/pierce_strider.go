package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pierce Strider — Artifact Creature — Phyrexian Construct {4}, 3/3 (Mirrodin Besieged,
// #5576):
//
//	"When this creature enters, target opponent loses 3 life."
//
// Geralf's Messenger's enters drain at 3: the opponent is chosen as the
// trigger goes on the stack (CR 603.3d) and loses the life if still legal
// as it resolves. Life loss, not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e2bbf963-794d-4600-bb92-cb9d1d00205a",
		Name:         "Pierce Strider",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Pierce Strider — target opponent loses 3 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return GainLife{Player: t.ID, Amount: -3}.Apply(ctx)
					}
					return nil
				}), TargetPlayer("target opponent", Opponent())),
		},
	})
}
