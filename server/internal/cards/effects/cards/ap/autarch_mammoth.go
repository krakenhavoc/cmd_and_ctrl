package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Autarch Mammoth — Creature — Elephant Mount {4}{G}{G}:
//
//	"When this creature enters and whenever it attacks while saddled,
//	 create a 3/3 green Elephant creature token.
//	 Saddle 5"
//
// One printed ability with two conditions, so two declarations sharing
// one body.
//
// No simplification.
func init() {
	const label = "Autarch Mammoth — create a 3/3 green Elephant"
	makeElephant := effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Elephant"), N: 1})
	effects.Register(effects.Spec{
		OracleID:     "eb50b5f2-4840-4f90-9c85-40fc581be9c7",
		Name:         "Autarch Mammoth",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Saddle(5)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, makeElephant),
			effects.AttacksWhileSaddled(label, makeElephant),
		},
	})
}
