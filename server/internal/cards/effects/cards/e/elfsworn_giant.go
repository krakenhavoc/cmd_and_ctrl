package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elfsworn Giant — Creature — Giant {3}{G}{G}, 5/3 (Foundations, #4465):
//
//	"Reach
//	 Landfall — Whenever a land you control enters, create a 1/1 green
//	 Elf Warrior creature token."
//
// Zendikar's Roil's landfall token on a reach body: one token for each
// land that enters under your control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f6abdf34-cfdc-465c-b067-cdebb659406e",
		Name:            "Elfsworn Giant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.LandfallPaying("Elfsworn Giant — create a 1/1 green Elf Warrior (landfall)", game.LandfallPayoff{Tokens: 1},
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Elf Warrior"), N: 1})),
		},
	})
}
