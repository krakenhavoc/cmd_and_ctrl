package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chimney Imp — Creature — Imp {4}{B}, 1/2 (Mirrodin, #5606):
//
//	"Flying
//	 When this creature dies, target opponent puts a card from their
//	 hand on top of their library."
//
// The targeted opponent chooses the card, through Volrath's Dungeon's
// put-back (PutFromHandOnTopInAnyOrder, N = 1); an opponent with an
// empty hand puts back nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3901bf30-b7c1-4977-a7b1-fcdafcc266cd",
		Name:            "Chimney Imp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Chimney Imp — target opponent puts a card from their hand on top of their library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return effects.PutFromHandOnTopInAnyOrder{
								Player: t.ID,
								N:      1,
								Label:  "Chimney Imp — put a card from your hand on top of your library",
							}.Apply(ctx)
						}
					}
					return nil
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
