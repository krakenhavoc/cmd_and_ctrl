package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Multani's Acolyte — Creature — Elf, {G}{G}, 2/1:
//
//	"Echo {G}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, draw a card."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d08725e6-ed9a-4c82-89d4-0dd037c376f6",
		Name:         "Multani's Acolyte",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Echo("Multani's Acolyte", "{G}{G}"),
			effects.WhenThisEnters("Multani's Acolyte — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
