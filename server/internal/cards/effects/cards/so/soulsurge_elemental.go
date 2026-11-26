package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soulsurge Elemental — {3}{R} Creature — Elemental */1 (Rise of the Eldrazi, S82 OLD #5581):
//
//	First strike
//	Soulsurge Elemental's power is equal to the number of creatures you control.
//
// The power is a characteristic-defining ability (Layer 7a, CR 604.3), counted as the layers run, so it follows creatures entering and leaving; the printed */1 is 0 power off the battlefield. It counts itself.

func init() {
	effects.Register(effects.Spec{
		OracleID:        "a98764e2-3cf6-40a4-aba5-34e2411623ab",
		Name:            "Soulsurge Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Static: []game.StaticAbility{
			effects.JmpB5PowerIsCreaturesYouControl("Soulsurge Elemental's power is equal to the number of creatures you control."),
		},
	})
}
