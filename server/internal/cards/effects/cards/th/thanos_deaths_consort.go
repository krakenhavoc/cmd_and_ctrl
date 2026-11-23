package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "3c067370-1e3d-4262-b7d6-f2bb2656b1d8",
		Name:            "Thanos, Death's Consort",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.AnotherCreatureDied,
				"Thanos — put a +1/+1 counter on Thanos",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
