package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sigiled Contender — Creature — Human Warrior {3}{W}, 3/3 (Core Set 2021, #5511):
//
//	"This creature has lifelink as long as it has a +1/+1 counter on it."
//
// Lightwalker's conditional keyword (ThisHasWhileItHasCounter) with
// lifelink: a layer 6 grant that holds only while a +1/+1 counter is on it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84ec7820-6608-459c-b172-80e8e3750cf4",
		Name:         "Sigiled Contender",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.ThisHasWhileItHasCounter(game.CounterPlusOne, "lifelink")},
	})
}
