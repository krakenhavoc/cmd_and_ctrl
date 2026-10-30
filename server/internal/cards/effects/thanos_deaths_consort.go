package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thanos, Death's Consort — Legendary Creature — Eternal Villain {2}{B}{B},
// 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Lifelink
//	 Whenever another creature dies, put a +1/+1 counter on Thanos."
//
// The any-creature death watcher (AnotherCreatureDied) with a self counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3c067370-1e3d-4262-b7d6-f2bb2656b1d8",
		Name:            "Thanos, Death's Consort",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, AnotherCreatureDied,
				"Thanos — put a +1/+1 counter on Thanos",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
