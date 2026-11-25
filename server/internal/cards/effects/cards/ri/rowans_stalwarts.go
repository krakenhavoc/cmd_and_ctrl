package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rowan's Stalwarts — Creature — Human Knight {4}{R}, 5/2 (Throne of
// Eldraine, #5516):
//
//	"When this creature enters, you may search your library and/or
//	 graveyard for a card named Rowan, Fearless Sparkmage, reveal it, and
//	 put it into your hand. If you search your library this way, shuffle."
//
// The partner-search trigger (MayFetchNamedFromLibraryOrGraveyard), as on
// Sorin's Guide.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9171248e-045e-4ed4-aa06-eea3ae3d8df0",
		Name:         "Rowan's Stalwarts",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.MayFetchNamedFromLibraryOrGraveyard("Rowan's Stalwarts", "Rowan, Fearless Sparkmage")},
	})
}
