package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Excavation Mole — Creature — Mole {2}{G}, 3/3 (Ikoria: Lair of Behemoths, #5512):
//
//	"Trample
//	 When this creature enters, mill three cards."
//
// Aftermath Analyst's entry mill.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "68e0b691-f332-4b0a-8173-bb8a48b463a1",
		Name:            "Excavation Mole",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Excavation Mole — mill three cards", effects.Do(effects.MillCards{N: 3})),
		},
	})
}
