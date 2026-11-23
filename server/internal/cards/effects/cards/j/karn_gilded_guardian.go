package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "20b90a18-ff62-4995-b0d6-fbef44dca357",
		Name:            "Karn, Gilded Guardian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Karn, Gilded Guardian — draw a card for each color among other artifacts you control", effects.RfCreatureCKarnDrawEffect),
		},
	})
}
