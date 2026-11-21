package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Anafenza, Kin-Tree Spirit — Legendary Creature — Spirit Soldier {W}{W},
// 2/2:
//
//	"Whenever another nontoken creature you control enters, bolster 1."
//
// Bolster is effects.Bolster (CR 701.39): the creature with the least
// toughness among those you control gets the counter, chosen as the
// trigger resolves, so the entering creature may itself be the one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7fa6ac0-70db-4586-b881-5a19db644dfb",
		Name:         "Anafenza, Kin-Tree Spirit",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AnotherNontokenCreatureEnteredUnderYourControl,
				"Anafenza, Kin-Tree Spirit — bolster 1",
				effects.Do(effects.Bolster{N: 1})),
		},
	})
}
