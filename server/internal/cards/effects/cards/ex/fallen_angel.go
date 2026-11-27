package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fallen Angel — Creature — Angel {3}{B}{B}, 3/3 (Commander 2011, #5574):
//
//	"Flying
//	 Sacrifice a creature: This creature gets +2/+1 until end of turn."
//
// The sacrifice is a cost paid as the ability is activated (CR 602.2), and it
// may be the Angel itself; with no mana in it the ability can be activated as
// often as there are creatures to give, each resolution adding another +2/+1
// until the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1869a326-8b48-4331-b969-5ce83cf4ff8a",
		Name:            "Fallen Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a creature: This creature gets +2/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 1}},
			Cost:    effects.SacrificeACreature(),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 1, "Fallen Angel — +2/+1"),
		}},
	})
}
