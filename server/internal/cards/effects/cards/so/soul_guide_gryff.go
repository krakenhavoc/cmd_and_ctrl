package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul-Guide Gryff — Creature — Hippogriff Spirit {4}{W}, 3/4 (Innistrad:
// Midnight Hunt, S82):
//
//	"Flying
//	 When this creature enters, exile up to one target card from a
//	 graveyard."
//
// Flying is the printed keyword. The target is "up to one": with a card
// chosen, it is re-checked as the trigger resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e26a08de-6108-4051-9c42-b5ce9dd8b533",
		Name:            "Soul-Guide Gryff",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{effects.S82MidB07ExileUpToOneFromAGraveyard("Soul-Guide Gryff")},
	})
}
