package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Windwright Mage — Artifact Creature — Human Wizard {W}{U}{B}, 2/2 (Shards
// of Alara, S82 OLD batch 512, #5588):
//
//	"Lifelink
//	 This creature has flying as long as an artifact card is in your
//	 graveyard."
func init() {
	flying := effects.KeywordGrant(effects.WhileYouHave(effects.S82OldB512ArtifactCardInYourGraveyard, effects.SelfCreature), "flying")
	flying.Label = "This creature has flying as long as an artifact card is in your graveyard."
	effects.Register(effects.Spec{
		OracleID:        "2f051ee4-9f2e-414d-b763-5fd78f2d4e1b",
		Name:            "Windwright Mage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Static:          []game.StaticAbility{flying},
	})
}
