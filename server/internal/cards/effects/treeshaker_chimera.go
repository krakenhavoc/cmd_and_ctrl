package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Treeshaker Chimera — Creature — Chimera, {5}{G}{G}, 8/5:
//
//	"All creatures able to block this creature do so.
//	 When this creature dies, draw three cards."
//
// #1684 leftover: the Lure clause is the Prized Unicorn shape
// (AllAbleToBlockDoSo); the dies trigger is WhenThisDies (graveyard
// only, CR 700.4) plus DrawCards. No new machinery.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b119b69-4794-443e-aa13-747c125b5b1b",
		Name:         "Treeshaker Chimera",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{AllAbleToBlockDoSo()},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Treeshaker Chimera — draw three cards", Do(DrawCards{N: 3})),
		},
	})
}
