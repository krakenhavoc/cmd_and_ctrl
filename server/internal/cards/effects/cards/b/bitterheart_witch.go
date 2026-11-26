package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bitterheart Witch — Creature — Human Shaman {4}{B}, 1/2 (Innistrad, S82
// older sets batch 449, #5572):
//
//	"Deathtouch
//	 When this creature dies, you may search your library for a Curse card,
//	 put it onto the battlefield attached to target player, then shuffle."
//
// The trigger targets a player (CR 603.3d), is optional, and the search puts
// the Curse onto the battlefield attached to that player (Boonweaver Giant's
// seam, CR 303.4i). The library is shuffled whether or not a card is found.
//
// No simplification.
func init() {
	witch := effects.Optional(
		effects.WhenThisDies("Bitterheart Witch — search your library for a Curse card, put it onto the battlefield attached to target player, then shuffle",
			effects.S82OldB449WitchSearch),
		"Bitterheart Witch — search your library for a Curse card?")
	witch.Targets = effects.TargetPlayer("target player")
	effects.Register(effects.Spec{
		OracleID:        "67466d56-9f28-4bbe-ae54-6fd834581e77",
		Name:            "Bitterheart Witch",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered:       []game.TriggeredAbility{witch},
	})
}
