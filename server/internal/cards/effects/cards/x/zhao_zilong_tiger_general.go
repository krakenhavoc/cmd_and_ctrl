package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zhao Zilong, Tiger General — Legendary Creature — Human Soldier Warrior {3}{W}{W}, 3/3
// (Portal Three Kingdoms, #5622):
//
//	"Horsemanship (This creature can't be blocked except by creatures with horsemanship.)"
//	"Whenever Zhao Zilong blocks, it gets +1/+1 until end of turn."
//
// Horsemanship is an engine keyword carried as printed data: the block
// check refuses any blocker without horsemanship. It doesn't stop this
// creature blocking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "22aa8e7f-b30d-49e2-b99e-9232e1a3bcf3",
		Name:            "Zhao Zilong, Tiger General",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SelfBlocksOnce(ev, source)
			}, "Zhao Zilong — gets +1/+1 until end of turn",
				effects.ThisGetsUntilEndOfTurn(1, 1, "Zhao Zilong — +1/+1")),
		},
	})
}
