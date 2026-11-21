package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saruman the White — Legendary Creature — Avatar Wizard {4}{U}, 4/4 (The
// Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Ward {2}
//	 Whenever you cast your second spell each turn, amass Orcs 2."
//
// Fractal Tender's ward {2} trigger and Alphinaud Leveilleur's second-spell
// watch, with Foray of Orcs's amass.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ba24644-8dff-4f1e-a96a-7361baf46b5a",
		Name:         "Saruman the White",
		Purpose:      game.Purpose{Amass: 2},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Saruman the White — ward {2}"),
			effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
				"Saruman the White — amass Orcs 2", effects.Do(effects.Amass{Subtype: "Orc", N: 2})),
		},
	})
}
