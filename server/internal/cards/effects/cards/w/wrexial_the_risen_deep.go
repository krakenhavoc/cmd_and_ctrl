package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wrexial, the Risen Deep — Legendary Creature — Kraken {3}{U}{U}{B}, 5/8
// (S82 reprint pass):
//
//	"Islandwalk, swampwalk
//	 Whenever Wrexial deals combat damage to a player, you may cast target
//	 instant or sorcery card from that player's graveyard without paying its
//	 mana cost. If that spell would be put into a graveyard, exile it
//	 instead."
//
// The trigger is Deluxe Dragster's (freeCastFromDamagedPlayersGraveyard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fd36cc16-d3d9-4c9d-9d28-bbe7e5459d75",
		Name:            "Wrexial, the Risen Deep",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"islandwalk", "swampwalk"},
		Triggered:       []game.TriggeredAbility{effects.FreeCastFromDamagedPlayersGraveyard("Wrexial, the Risen Deep")},
	})
}
