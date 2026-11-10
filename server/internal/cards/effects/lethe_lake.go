package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lethe Lake — Plane — Arkhos (#4870):
//
//	"At the beginning of your upkeep, mill ten cards.
//	 Whenever chaos ensues, target player mills ten cards."
//
// "Your upkeep" is the planar controller's. A library with fewer than ten
// cards mills what it has and nobody loses for it (CR 701.17b, the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "85592710-8637-4726-8b92-552b9be677c0",
		Name:         "Lethe Lake",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Lethe Lake — mill ten cards", Do(MillCards{N: 10})),
			Targeting(WheneverChaosEnsues("Lethe Lake — target player mills ten cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return MillCards{Player: t.ID, N: 10}.Apply(ctx)
				}),
				TargetPlayer("target player")),
		},
	})
}
