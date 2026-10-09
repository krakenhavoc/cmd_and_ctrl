package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Karn, Gilded Guardian — Legendary Artifact Creature — Golem
// {2/W}{2/U}{2/B}{2/R}{2/G}, 5/5:
//
//	"Vigilance, trample
//	 When Karn enters, draw a card for each color among other artifacts
//	 you control."
//
// Each distinct color among the controller's OTHER artifacts counts once
// (Karn himself is colored, and is excluded). The colors are read as the
// ability resolves, from the effective characteristics, so a color-
// changed artifact counts as what it is now. The five hybrid symbols are
// the printed cost; Scryfall's mana cost is what the engine charges.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "20b90a18-ff62-4995-b0d6-fbef44dca357",
		Name:            "Karn, Gilded Guardian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Karn, Gilded Guardian — draw a card for each color among other artifacts you control", rfCreatureCKarnDrawEffect),
		},
	})
}
