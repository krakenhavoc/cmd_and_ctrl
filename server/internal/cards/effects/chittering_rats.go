package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chittering Rats — Creature — Rat {1}{B}{B} (Darksteel, #5605):
//
//	"When this creature enters, target opponent puts a card from their
//	 hand on top of their library."
//
// Chimney Imp's effect on an enters trigger: the opponent picks the card
// (PutFromHandOnTopInAnyOrder), and an empty hand puts nothing back.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "08dfe42e-35c0-4be0-abba-57269792ff3d",
		Name:         "Chittering Rats",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Chittering Rats — target opponent puts a card from their hand on top of their library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return PutFromHandOnTopInAnyOrder{
								Player: t.ID,
								N:      1,
								Label:  "Chittering Rats — put a card from your hand on top of your library",
							}.Apply(ctx)
						}
					}
					return nil
				}),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
