package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dori, Bearer of Friends — Legendary Creature — Dwarf Warrior {2}{R}, 3/2
// (The Hobbit, HOB):
//
//	"Trample
//	 When Dori enters, create a Treasure token."
//
// No simplification (#4454).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "58acd6b2-730c-4472-a34a-f2784977aca8",
		Name:            "Dori, Bearer of Friends",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dori, Bearer of Friends — create a Treasure token", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}
