package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skemfar Avenger — Creature — Elf Berserker {1}{B}, 3/1 (Kaldheim):
//
//	"Whenever another nontoken Elf or Berserker you control dies, you draw a
//	 card and you lose 1 life."
//
// The dead creature is read as it last existed (CR 603.10a): its creature
// types at that moment decide Elf or Berserker, and a token is not counted.
// One trigger, so a creature that is both an Elf and a Berserker draws once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "47c67df1-52f5-459f-b83e-6b971be9b294",
		Name:         "Skemfar Avenger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.S82OldB061NontokenElfOrBerserkerYouControlDied,
				"Skemfar Avenger — you draw a card and you lose 1 life",
				effects.Do(effects.DrawCards{N: 1}, effects.YouLoseLife(1))),
		},
	})
}
