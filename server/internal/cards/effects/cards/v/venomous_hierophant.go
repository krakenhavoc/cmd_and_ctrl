package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Venomous Hierophant — Creature — Gorgon Cleric {3}{B}, 3/3 (Ravnica: Clue
// Edition, S82):
//
//	"Deathtouch
//	 When this creature enters, mill three cards."
//
// Colossal Grave-Reaver's mill body on an enters trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e235d1ca-3099-46c1-8f33-798286292d71",
		Name:            "Venomous Hierophant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Venomous Hierophant — mill three cards", effects.Do(effects.MillCards{N: 3})),
		},
	})
}
