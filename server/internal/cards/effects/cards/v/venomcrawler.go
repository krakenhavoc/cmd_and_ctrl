package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Venomcrawler — Artifact Creature — Demon {3}{B}, 2/2:
//
//	"Lifelink
//	 Devourer of Souls — Whenever another creature dies, put a +1/+1 counter
//	 on this creature."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f3a07ae4-b0f8-43e6-adba-217f220cc2d0",
		Name:            "Venomcrawler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.AnotherCreatureDied, "Venomcrawler — put a +1/+1 counter on this creature", effects.PlusOneCountersOnThis(1)),
		},
	})
}
