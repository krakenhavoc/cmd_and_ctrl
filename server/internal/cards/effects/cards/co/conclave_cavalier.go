package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conclave Cavalier — Creature — Centaur Knight {G}{G}{W}{W}, 4/4 (Ravnica
// Remastered, #5948):
//
//	"Vigilance
//	 When this creature dies, create two 2/2 green and white Elf Knight
//	 creature tokens with vigilance."
//
// Vigilance is the engine keyword. The tokens are made by the controller the
// Cavalier had as it died (last-known information, CR 603.10a); the token row is
// the printed Elf Knight token (see s82_gap_b73_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "653de668-beb0-4be7-bb05-30de5f52a3f1",
		Name:            "Conclave Cavalier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Conclave Cavalier — create two 2/2 green and white Elf Knight creature tokens with vigilance",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green and white Elf Knight with vigilance"), N: 2})),
		},
	})
}
