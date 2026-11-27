package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Diregraf Escort — {G} Creature — Human Cleric 1/1: "Soulbond. As long as this
// creature is paired with another creature, both creatures have protection from
// Zombies."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f28a67d8-05ec-4d72-aa17-670ce1136f5c",
		Name:            "Diregraf Escort",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered:       effects.Soulbond("Diregraf Escort"),
		Static: []game.StaticAbility{
			effects.PairedCreaturesHaveKeyword("As long as this creature is paired with another creature, both creatures have protection from Zombies.", "protection from Zombies"),
		},
	})
}
